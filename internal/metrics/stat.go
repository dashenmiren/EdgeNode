package metrics

import (
	"strconv"
	"strings"

	"github.com/dashenmiren/EdgeNode/internal/utils/fnv"
)

type Stat struct {
	ServerId int64
	Keys     []string
	Hash     string
	Value    int64
	Time     string
}

func SumStat(serverId int64, keys []string, time string, version int32, itemId int64) string {
	var keysData = strings.Join(keys, "$EDGE$")
	return strconv.FormatUint(fnv.HashString(strconv.FormatInt(serverId, 10)+"@"+keysData+"@"+time+"@"+strconv.Itoa(int(version))+"@"+strconv.FormatInt(itemId, 10)), 10)
}
