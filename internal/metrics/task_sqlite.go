// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package metrics

import (
	"encoding/json"
	"github.com/dashenmiren/EdgeCommon/pkg/rpc/pb"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	"github.com/dashenmiren/EdgeNode/internal/rpc"
	"github.com/dashenmiren/EdgeNode/internal/utils"
	"github.com/dashenmiren/EdgeNode/internal/utils/dbs"
	"github.com/dashenmiren/EdgeNode/internal/utils/goman"
	"github.com/dashenmiren/EdgeNode/internal/utils/idles"
	"github.com/dashenmiren/EdgeNode/internal/utils/trackers"
	"github.com/dashenmiren/EdgeNode/internal/utils/zero"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/types"
	"os"
	"strconv"
	"sync"
	"time"
)

const MaxQueueSize = 256 // TODO 可以配置，可以在单个任务里配置

// SQLiteTask 单个指标任务
// 数据库存储：
//
//	data/
//	   metric.$ID.db
//	      stats
//	         id, keys, value, time, serverId, hash
//	原理：
//	   添加或者有变更时 isUploaded = false
//	   上传时检查 isUploaded 状态
//	   只上传每个服务中排序最前面的 N 个数据
type SQLiteTask struct {
	BaseTask

	db            *dbs.DB
	statTableName string

	cleanTicker  *time.Ticker
	uploadTicker *utils.Ticker

	cleanVersion int32

	insertStatStmt          *dbs.Stmt
	deleteByVersionStmt     *dbs.Stmt
	deleteByExpiresTimeStmt *dbs.Stmt
	selectTopStmt           *dbs.Stmt
	sumStmt                 *dbs.Stmt

	serverIdMap       map[int64]zero.Zero  // 所有的服务Ids
	timeMap           map[string]zero.Zero // time => bool
	serverIdMapLocker sync.Mutex

	statsTicker *utils.Ticker
}

// NewSQLiteTask 获取新任务
func NewSQLiteTask(item *serverconfigs.MetricItemConfig) *SQLiteTask {
	return &SQLiteTask{
		BaseTask: BaseTask{
			itemConfig: item,
			statsMap:   map[string]*Stat{},
		},

		serverIdMap: map[int64]zero.Zero{},
		timeMap:     map[string]zero.Zero{},
	}
}

func CheckSQLiteDB(itemId int64) bool {
	var path = Tea.Root + "/data/metric." + types.String(itemId) + ".db"
	_, err := os.Stat(path)
	return err == nil
}

// Init 初始化
func (this *SQLiteTask) Init() error {
	this.statTableName = "stats"

	// 检查目录是否存在
	var dir = Tea.Root + "/data"
	_, err := os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
		remotelogs.Println("METRIC", "create data dir '"+dir+"'")
	}

	var path = dir + "/metric." + types.String(this.itemConfig.Id) + ".db"

	db, err := dbs.OpenWriter("file:" + path + "?cache=shared&mode=rwc&_journal_mode=WAL&_sync=" + dbs.SyncMode + "&_locking_mode=EXCLUSIVE")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	this.db = db

	// 恢复数据库
	var recoverEnv, _ = os.LookupEnv("EdgeRecover")
	if len(recoverEnv) > 0 {
		for _, indexName := range []string{"serverId", "hash"} {
			_, _ = db.Exec(`REINDEX "` + indexName + `"`)
		}
	}

	if teaconst.EnableDBStat {
		this.db.EnableStat(true)
	}

	//创建统计表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "` + this.statTableName + `" (
  "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  "hash" varchar(32),
  "keys" varchar(1024),
  "value" real DEFAULT 0,
  "time" varchar(32),
  "serverId" integer DEFAULT 0,
  "version" integer DEFAULT 0,
  "isUploaded" integer DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "serverId"
ON "` + this.statTableName + `" (
  "serverId" ASC,
  "version" ASC
);

CREATE UNIQUE INDEX IF NOT EXISTS "hash"
ON "` + this.statTableName + `" (
  "hash" ASC
);`)
	if err != nil {
		return err
	}

	// insert stat stmt
	this.insertStatStmt, err = db.Prepare(`INSERT INTO "stats" ("serverId", "hash", "keys", "value", "time", "version", "isUploaded") VALUES (?, ?, ?, ?, ?, ?, 0) ON CONFLICT("hash") DO UPDATE SET "value"="value"+?, "isUploaded"=0`)
	if err != nil {
		return err
	}

	// delete by version
	this.deleteByVersionStmt, err = db.Prepare(`DELETE FROM "` + this.statTableName + `" WHERE "version"<?`)
	if err != nil {
		return err
	}

	// delete by expires time
	this.deleteByExpiresTimeStmt, err = db.Prepare(`DELETE FROM "` + this.statTableName + `" WHERE "time"<?`)
	if err != nil {
		return err
	}

	// select topN stmt
	this.selectTopStmt, err = db.Prepare(`SELECT "id", "hash", "keys", "value", "isUploaded" FROM "` + this.statTableName + `" WHERE "serverId"=? AND "version"=? AND time=? ORDER BY "value" DESC LIMIT 20`)
	if err != nil {
		return err
	}

	// sum stmt
	this.sumStmt, err = db.Prepare(`SELECT COUNT(*), IFNULL(SUM(value), 0) FROM "` + this.statTableName + `" WHERE "serverId"=? AND "version"=? AND time=?`)
	if err != nil {
		return err
	}

	// 所有的服务IDs
	err = this.loadServerIdMap()
	if err != nil {
		return err
	}

	this.isLoaded = true

	return nil
}

// Start 启动任务
func (this *SQLiteTask) Start() error {
	// 读取数据
	this.statsTicker = utils.NewTicker(1 * time.Minute)
	goman.New(func() {
		for this.statsTicker.Next() {
			var tr = trackers.Begin("METRIC:DUMP_STATS_TO_LOCAL_DATABASE")

			this.statsLocker.Lock()
			var statsMap = this.statsMap
			this.statsMap = map[string]*Stat{}
			this.statsLocker.Unlock()

			for _, stat := range statsMap {
				err := this.InsertStat(stat)
				if err != nil {
					remotelogs.Error("METRIC", "insert stat failed: "+err.Error())
				}
			}

			tr.End()
		}
	})

	// 清理
	this.cleanTicker = time.NewTicker(24 * time.Hour)
	goman.New(func() {
		idles.RunTicker(this.cleanTicker, func() {
			var tr = trackers.Begin("METRIC:CLEAN_EXPIRED")
			err := this.CleanExpired()
			tr.End()
			if err != nil {
				remotelogs.Error("METRIC", "clean expired stats failed: "+err.Error())
			}
		})
	})

	// 上传
	this.uploadTicker = utils.NewTicker(this.itemConfig.UploadDuration())
	goman.New(func() {
		for this.uploadTicker.Next() {
			err := this.Upload(1 * time.Second)
			if err != nil && !rpc.IsConnError(err) {
				remotelogs.Error("METRIC", "upload stats failed: "+err.Error())
			}
		}
	})

	return nil
}

// Stop 停止任务
func (this *SQLiteTask) Stop() error {
	this.isStopped = true

	if this.cleanTicker != nil {
		this.cleanTicker.Stop()
	}
	if this.uploadTicker != nil {
		this.uploadTicker.Stop()
	}
	if this.statsTicker != nil {
		this.statsTicker.Stop()
	}

	_ = this.insertStatStmt.Close()
	_ = this.deleteByVersionStmt.Close()
	_ = this.deleteByExpiresTimeStmt.Close()
	_ = this.selectTopStmt.Close()
	_ = this.sumStmt.Close()

	if this.db != nil {
		_ = this.db.Close()
	}

	return nil
}

func (this *SQLiteTask) Delete() error {
	return nil
}

// InsertStat 写入数据
func (this *SQLiteTask) InsertStat(stat *Stat) error {
	if this.isStopped {
		return nil
	}
	if stat == nil {
		return nil
	}

	this.serverIdMapLocker.Lock()
	this.serverIdMap[stat.ServerId] = zero.New()
	this.timeMap[stat.Time] = zero.New()
	this.serverIdMapLocker.Unlock()

	keyData, err := json.Marshal(stat.Keys)
	if err != nil {
		return err
	}

	_, err = this.insertStatStmt.Exec(stat.ServerId, stat.Hash, keyData, stat.Value, stat.Time, this.itemConfig.Version, stat.Value)
	if err != nil {
		return err
	}
	return nil
}

// CleanExpired 清理数据
func (this *SQLiteTask) CleanExpired() error {
	if this.isStopped {
		return nil
	}

	// 清除低版本数据
	if this.cleanVersion < this.itemConfig.Version {
		_, err := this.deleteByVersionStmt.Exec(this.itemConfig.Version)
		if err != nil {
			return err
		}
		this.cleanVersion = this.itemConfig.Version
	}

	// 清除过期的数据
	_, err := this.deleteByExpiresTimeStmt.Exec(this.itemConfig.LocalExpiresTime())
	if err != nil {
		return err
	}

	return nil
}

// Upload 上传数据
func (this *SQLiteTask) Upload(pauseDuration time.Duration) error {
	var uploadTr = trackers.Begin("METRIC:UPLOAD_STATS")
	defer uploadTr.End()

	if this.isStopped {
		return nil
	}

	this.serverIdMapLocker.Lock()

	// 服务IDs
	var serverIds []int64
	for serverId := range this.serverIdMap {
		serverIds = append(serverIds, serverId)
	}
	this.serverIdMap = map[int64]zero.Zero{} // 清空数据

	// 时间
	var times = []string{}
	for t := range this.timeMap {
		times = append(times, t)
	}
	this.timeMap = map[string]zero.Zero{} // 清空数据

	this.serverIdMapLocker.Unlock()

	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		return err
	}

	for _, serverId := range serverIds {
		for _, currentTime := range times {
			idStrings, err := func(serverId int64, currentTime string) (ids []string, err error) {
				var t = trackers.Begin("METRIC:SELECT_TOP_STMT")
				rows, err := this.selectTopStmt.Query(serverId, this.itemConfig.Version, currentTime)
				t.End()
				if err != nil {
					return nil, err
				}
				var isClosed bool
				defer func() {
					if isClosed {
						return
					}
					_ = rows.Close()
				}()

				var pbStats []*pb.UploadingMetricStat
				for rows.Next() {
					var pbStat = &pb.UploadingMetricStat{}
					// "id", "hash", "keys", "value", "isUploaded"
					var isUploaded int
					var keysData []byte
					err = rows.Scan(&pbStat.Id, &pbStat.Hash, &keysData, &pbStat.Value, &isUploaded)
					if err != nil {
						return nil, err
					}

					// TODO 先不判断是否已经上传，需要改造API进行配合
					/**if isUploaded == 1 {
						continue
					}**/
					if len(keysData) > 0 {
						err = json.Unmarshal(keysData, &pbStat.Keys)
						if err != nil {
							return nil, err
						}
					}
					pbStats = append(pbStats, pbStat)
					ids = append(ids, strconv.FormatInt(pbStat.Id, 10))
				}

				// 提前关闭
				_ = rows.Close()
				isClosed = true

				// 上传
				if len(pbStats) > 0 {
					// 计算总和
					count, total, err := this.sum(serverId, currentTime)
					if err != nil {
						return nil, err
					}

					_, err = rpcClient.MetricStatRPC.UploadMetricStats(rpcClient.Context(), &pb.UploadMetricStatsRequest{
						MetricStats: pbStats,
						Time:        currentTime,
						ServerId:    serverId,
						ItemId:      this.itemConfig.Id,
						Version:     this.itemConfig.Version,
						Count:       count,
						Total:       float32(total),
					})
					if err != nil {
						return nil, err
					}
				}

				return
			}(serverId, currentTime)
			if err != nil {
				return err
			}

			if len(idStrings) > 0 {
				// 设置为已上传
				// TODO 先不判断是否已经上传，需要改造API进行配合
				/**_, err = this.db.Exec(`UPDATE "` + this.statTableName + `" SET isUploaded=1 WHERE id IN (` + strings.Join(idStrings, ",") + `)`)
				if err != nil {
					return err
				}**/
			}
		}

		// 休息一下，防止短时间内上传数据过多
		if pauseDuration > 0 {
			time.Sleep(pauseDuration)
			uploadTr.Add(-pauseDuration)
		}
	}

	return nil
}

// 加载服务ID
func (this *SQLiteTask) loadServerIdMap() error {
	{
		rows, err := this.db.Query(`SELECT DISTINCT "serverId" FROM `+this.statTableName+" WHERE version=?", this.itemConfig.Version)
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
		}()

		var serverId int64
		for rows.Next() {
			err = rows.Scan(&serverId)
			if err != nil {
				return err
			}
			this.serverIdMapLocker.Lock()
			this.serverIdMap[serverId] = zero.New()
			this.serverIdMapLocker.Unlock()
		}
	}

	{
		rows, err := this.db.Query(`SELECT DISTINCT "time" FROM `+this.statTableName+" WHERE version=?", this.itemConfig.Version)
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
		}()

		var timeString string
		for rows.Next() {
			err = rows.Scan(&timeString)
			if err != nil {
				return err
			}
			this.serverIdMapLocker.Lock()
			this.timeMap[timeString] = zero.New()
			this.serverIdMapLocker.Unlock()
		}
	}

	return nil
}

// 计算数量和综合
func (this *SQLiteTask) sum(serverId int64, time string) (count int64, total float64, err error) {
	rows, err := this.sumStmt.Query(serverId, this.itemConfig.Version, time)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()
	if rows.Next() {
		err = rows.Scan(&count, &total)
		if err != nil {
			return 0, 0, err
		}
	}
	return
}
