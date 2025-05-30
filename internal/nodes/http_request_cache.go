package nodes

import (
	"bytes"
	"errors"
	"github.com/dashenmiren/EdgeCommon/pkg/configutils"
	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/dashenmiren/EdgeNode/internal/compressions"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	"github.com/dashenmiren/EdgeNode/internal/utils"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	rangeutils "github.com/dashenmiren/EdgeNode/internal/utils/ranges"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 读取缓存
func (this *HTTPRequest) doCacheRead(useStale bool) (shouldStop bool) {
	// 需要动态Upgrade的不缓存
	if len(this.RawReq.Header.Get("Upgrade")) > 0 {
		return
	}

	this.cacheCanTryStale = false

	var cachePolicy = this.ReqServer.HTTPCachePolicy
	if cachePolicy == nil || !cachePolicy.IsOn {
		return
	}

	if this.web.Cache == nil || !this.web.Cache.IsOn || (len(cachePolicy.CacheRefs) == 0 && len(this.web.Cache.CacheRefs) == 0) {
		return
	}

	// 添加 X-Cache Header
	var addStatusHeader = this.web.Cache.AddStatusHeader
	var cacheBypassDescription = ""
	if addStatusHeader {
		defer func() {
			if len(cacheBypassDescription) > 0 {
				this.writer.Header().Set("X-Cache", cacheBypassDescription)
				return
			}
			var cacheStatus = this.varMapping["cache.status"]
			if cacheStatus != "HIT" {
				this.writer.Header().Set("X-Cache", cacheStatus)
			}
		}()
	}

	// 检查服务独立的缓存条件
	var refType = ""
	for _, cacheRef := range this.web.Cache.CacheRefs {
		if !cacheRef.IsOn {
			continue
		}
		if (cacheRef.Conds != nil && cacheRef.Conds.HasRequestConds() && cacheRef.Conds.MatchRequest(this.Format)) ||
			(cacheRef.SimpleCond != nil && cacheRef.SimpleCond.Match(this.Format)) {
			if cacheRef.IsReverse {
				return
			}
			this.cacheRef = cacheRef
			refType = "server"
			break
		}
	}
	if this.cacheRef == nil && !this.web.Cache.DisablePolicyRefs {
		// 检查策略默认的缓存条件
		for _, cacheRef := range cachePolicy.CacheRefs {
			if !cacheRef.IsOn {
				continue
			}
			if (cacheRef.Conds != nil && cacheRef.Conds.HasRequestConds() && cacheRef.Conds.MatchRequest(this.Format)) ||
				(cacheRef.SimpleCond != nil && cacheRef.SimpleCond.Match(this.Format)) {
				if cacheRef.IsReverse {
					return
				}
				this.cacheRef = cacheRef
				refType = "policy"
				break
			}
		}
	}

	if this.cacheRef == nil {
		return
	}

	// 是否强制Range回源
	if this.cacheRef.AlwaysForwardRangeRequest && len(this.RawReq.Header.Get("Range")) > 0 {
		this.cacheRef = nil
		cacheBypassDescription = "BYPASS, forward range"
		return
	}

	// 是否正在Purge
	var isPurging = this.web.Cache.PurgeIsOn && strings.ToUpper(this.RawReq.Method) == "PURGE" && this.RawReq.Header.Get("X-Edge-Purge-Key") == this.web.Cache.PurgeKey
	if isPurging {
		this.RawReq.Method = http.MethodGet
	}

	// 校验请求
	if !this.cacheRef.MatchRequest(this.RawReq) {
		this.cacheRef = nil
		cacheBypassDescription = "BYPASS, not match"
		return
	}

	// 相关变量
	this.varMapping["cache.policy.name"] = cachePolicy.Name
	this.varMapping["cache.policy.id"] = strconv.FormatInt(cachePolicy.Id, 10)
	this.varMapping["cache.policy.type"] = cachePolicy.Type

	// Cache-Pragma
	if this.cacheRef.EnableRequestCachePragma {
		if this.RawReq.Header.Get("Cache-Control") == "no-cache" || this.RawReq.Header.Get("Pragma") == "no-cache" {
			this.cacheRef = nil
			cacheBypassDescription = "BYPASS, Cache-Control or Pragma"
			return
		}
	}

	// TODO 支持Vary Header

	// 缓存标签
	var tags = []string{}

	// 检查是否有缓存
	var key string
	if this.web.Cache.Key != nil && this.web.Cache.Key.IsOn && len(this.web.Cache.Key.Host) > 0 {
		key = configutils.ParseVariables(this.cacheRef.Key, func(varName string) (value string) {
			switch varName {
			case "scheme":
				return this.web.Cache.Key.Scheme
			case "host":
				return this.web.Cache.Key.Host
			default:
				return this.Format("${" + varName + "}")
			}
		})
	} else {
		key = this.Format(this.cacheRef.Key)
	}

	if len(key) == 0 {
		this.cacheRef = nil
		cacheBypassDescription = "BYPASS, empty key"
		return
	}
	var method = this.Method()
	if method != http.MethodGet {
		key += caches.SuffixMethod + method
		tags = append(tags, strings.ToLower(method))
	}

	this.cacheKey = key
	this.varMapping["cache.key"] = key

	// 读取缓存
	var storage = caches.SharedManager.FindStorageWithPolicy(cachePolicy.Id)
	if storage == nil {
		this.cacheRef = nil
		cacheBypassDescription = "BYPASS, no policy found"
		return
	}
	this.writer.cacheStorage = storage

	// 如果正在预热，则不读取缓存，等待下一个步骤重新生成
	if (strings.HasPrefix(this.RawReq.RemoteAddr, "127.") || strings.HasPrefix(this.RawReq.RemoteAddr, "[::1]")) && this.RawReq.Header.Get("X-Edge-Cache-Action") == "fetch" {
		return
	}

	// 判断是否在Purge
	if isPurging {
		this.varMapping["cache.status"] = "PURGE"

		var subKeys = []string{
			key,
			key + caches.SuffixMethod + "HEAD",
			key + caches.SuffixWebP,
			key + caches.SuffixPartial,
		}
		// TODO 根据实际缓存的内容进行组合
		for _, encoding := range compressions.AllEncodings() {
			subKeys = append(subKeys, key+caches.SuffixCompression+encoding)
			subKeys = append(subKeys, key+caches.SuffixWebP+caches.SuffixCompression+encoding)
		}
		for _, subKey := range subKeys {
			err := storage.Delete(subKey)
			if err != nil {
				remotelogs.ErrorServer("HTTP_REQUEST_CACHE", "purge failed: "+err.Error())
			}
		}

		// 通过API节点清除别节点上的的Key
		SharedHTTPCacheTaskManager.PushTaskKeys([]string{key})

		return true
	}

	// 调用回调
	this.onRequest()
	if this.writer.isFinished {
		return
	}

	var reader caches.Reader
	var err error

	var rangeHeader = this.RawReq.Header.Get("Range")
	var isPartialRequest = len(rangeHeader) > 0

	// 检查是否支持WebP
	var webPIsEnabled = false
	var isHeadMethod = method == http.MethodHead
	if !isPartialRequest &&
		!isHeadMethod &&
		this.web.WebP != nil &&
		this.web.WebP.IsOn &&
		this.web.WebP.MatchRequest(filepath.Ext(this.Path()), this.Format) &&
		this.web.WebP.MatchAccept(this.RawReq.Header.Get("Accept")) {
		webPIsEnabled = true
	}

	// 检查WebP压缩缓存
	if webPIsEnabled && !isPartialRequest && !isHeadMethod && reader == nil {
		if this.web.Compression != nil && this.web.Compression.IsOn {
			_, encoding, ok := this.web.Compression.MatchAcceptEncoding(this.RawReq.Header.Get("Accept-Encoding"))
			if ok {
				reader, err = storage.OpenReader(key+caches.SuffixWebP+caches.SuffixCompression+encoding, useStale, false)
				if err != nil && caches.IsBusyError(err) {
					this.varMapping["cache.status"] = "BUSY"
					this.cacheRef = nil
					return
				}
				if reader != nil {
					tags = append(tags, "webp", encoding)
				}
			}
		}
	}

	// 检查WebP
	if webPIsEnabled && !isPartialRequest &&
		!isHeadMethod &&
		reader == nil {
		reader, err = storage.OpenReader(key+caches.SuffixWebP, useStale, false)
		if err != nil && caches.IsBusyError(err) {
			this.varMapping["cache.status"] = "BUSY"
			this.cacheRef = nil
			return
		}
		if reader != nil {
			this.writer.cacheReaderSuffix = caches.SuffixWebP
			tags = append(tags, "webp")
		}
	}

	// 检查普通压缩缓存
	if !isPartialRequest && !isHeadMethod && reader == nil {
		if this.web.Compression != nil && this.web.Compression.IsOn {
			_, encoding, ok := this.web.Compression.MatchAcceptEncoding(this.RawReq.Header.Get("Accept-Encoding"))
			if ok {
				reader, err = storage.OpenReader(key+caches.SuffixCompression+encoding, useStale, false)
				if err != nil && caches.IsBusyError(err) {
					this.varMapping["cache.status"] = "BUSY"
					this.cacheRef = nil
					return
				}
				if reader != nil {
					tags = append(tags, encoding)
				}
			}
		}
	}

	// 检查正常的文件
	var isPartialCache = false
	var partialRanges []rangeutils.Range
	var firstRangeEnd int64
	if reader == nil {
		reader, err = storage.OpenReader(key, useStale, false)
		if err != nil && caches.IsBusyError(err) {
			this.varMapping["cache.status"] = "BUSY"
			this.cacheRef = nil
			return
		}
		if err != nil && this.cacheRef.AllowPartialContent {
			// 尝试读取分片的缓存内容
			if len(rangeHeader) == 0 && this.cacheRef.ForcePartialContent {
				// 默认读取开头
				rangeHeader = "bytes=0-"
			}

			if len(rangeHeader) > 0 {
				pReader, ranges, rangeEnd, goNext := this.tryPartialReader(storage, key, useStale, rangeHeader, this.cacheRef.ForcePartialContent)
				if !goNext {
					this.cacheRef = nil
					return
				}
				if pReader != nil {
					isPartialCache = true
					reader = pReader
					partialRanges = ranges
					firstRangeEnd = rangeEnd
					err = nil
				}
			}
		}

		if err != nil {
			if errors.Is(err, caches.ErrNotFound) {
				// 移除请求中的 If-None-Match 和 If-Modified-Since，防止源站返回304而无法缓存
				if this.reverseProxy != nil {
					this.RawReq.Header.Del("If-None-Match")
					this.RawReq.Header.Del("If-Modified-Since")
				}

				// cache相关变量
				this.varMapping["cache.status"] = "MISS"

				if !useStale && this.web.Cache.Stale != nil && this.web.Cache.Stale.IsOn {
					this.cacheCanTryStale = true
				}
				return
			}

			if !this.canIgnore(err) {
				remotelogs.WarnServer("HTTP_REQUEST_CACHE", this.URL()+": read from cache failed: open cache failed: "+err.Error())
			}
			return
		}
	}

	defer func() {
		if !this.writer.DelayRead() {
			_ = reader.Close()
		}
	}()

	if useStale {
		this.varMapping["cache.status"] = "STALE"
		this.logAttrs["cache.status"] = "STALE"
	} else {
		this.varMapping["cache.status"] = "HIT"
		this.logAttrs["cache.status"] = "HIT"
	}

	// 准备Buffer
	var fileSize = reader.BodySize()
	var totalSizeString = types.String(fileSize)
	if isPartialCache {
		fileSize = reader.(*caches.PartialFileReader).MaxLength()
		if totalSizeString == "0" {
			totalSizeString = "*"
		}
	}

	// 读取Header
	var headerData = []byte{}
	this.writer.SetSentHeaderBytes(reader.HeaderSize())
	var headerPool = this.bytePool(reader.HeaderSize())
	var headerBuf = headerPool.Get()
	err = reader.ReadHeader(headerBuf.Bytes, func(n int) (goNext bool, readErr error) {
		headerData = append(headerData, headerBuf.Bytes[:n]...)
		for {
			var nIndex = bytes.Index(headerData, []byte{'\n'})
			if nIndex >= 0 {
				var row = headerData[:nIndex]
				var spaceIndex = bytes.Index(row, []byte{':'})
				if spaceIndex <= 0 {
					return false, errors.New("invalid header '" + string(row) + "'")
				}

				this.writer.Header().Set(string(row[:spaceIndex]), string(row[spaceIndex+1:]))
				headerData = headerData[nIndex+1:]
			} else {
				break
			}
		}
		return true, nil
	})
	headerPool.Put(headerBuf)
	if err != nil {
		if !this.canIgnore(err) {
			remotelogs.WarnServer("HTTP_REQUEST_CACHE", this.URL()+": read from cache failed: read header failed: "+err.Error())
		}
		return
	}

	// 设置cache.age变量
	var age = strconv.FormatInt(fasttime.Now().Unix()-reader.LastModified(), 10)
	this.varMapping["cache.age"] = age

	if addStatusHeader {
		if useStale {
			this.writer.Header().Set("X-Cache", "STALE, "+refType+", "+reader.TypeName())
		} else {
			this.writer.Header().Set("X-Cache", "HIT, "+refType+", "+reader.TypeName())
		}
	} else {
		this.writer.Header().Del("X-Cache")
	}
	if this.web.Cache.AddAgeHeader {
		this.writer.Header().Set("Age", age)
	}

	// ETag
	var respHeader = this.writer.Header()
	var eTag = respHeader.Get("ETag")
	var lastModifiedAt = reader.LastModified()
	if len(eTag) == 0 {
		if lastModifiedAt > 0 {
			if len(tags) > 0 {
				eTag = "\"" + strconv.FormatInt(lastModifiedAt, 10) + "_" + strings.Join(tags, "_") + "\""
			} else {
				eTag = "\"" + strconv.FormatInt(lastModifiedAt, 10) + "\""
			}
			respHeader.Del("Etag")
			if !isPartialCache {
				respHeader["ETag"] = []string{eTag}
			}
		}
	}

	// 支持 Last-Modified
	var modifiedTime = ""
	if lastModifiedAt > 0 {
		modifiedTime = time.Unix(utils.GMTUnixTime(lastModifiedAt), 0).Format("Mon, 02 Jan 2006 15:04:05") + " GMT"
		if !isPartialCache {
			respHeader.Set("Last-Modified", modifiedTime)
		}
	}

	// 支持 If-None-Match
	if !this.isLnRequest && !isPartialCache && len(eTag) > 0 && this.requestHeader("If-None-Match") == eTag {
		// 自定义Header
		this.ProcessResponseHeaders(this.writer.Header(), http.StatusNotModified)
		this.addExpiresHeader(reader.ExpiresAt())
		this.writer.WriteHeader(http.StatusNotModified)
		this.isCached = true
		this.cacheRef = nil
		this.writer.SetOk()
		return true
	}

	// 支持 If-Modified-Since
	if !this.isLnRequest && !isPartialCache && len(modifiedTime) > 0 && this.requestHeader("If-Modified-Since") == modifiedTime {
		// 自定义Header
		this.ProcessResponseHeaders(this.writer.Header(), http.StatusNotModified)
		this.addExpiresHeader(reader.ExpiresAt())
		this.writer.WriteHeader(http.StatusNotModified)
		this.isCached = true
		this.cacheRef = nil
		this.writer.SetOk()
		return true
	}

	this.ProcessResponseHeaders(this.writer.Header(), reader.Status())
	this.addExpiresHeader(reader.ExpiresAt())

	// 返回上级节点过期时间
	if this.isLnRequest {
		respHeader.Set(LNExpiresHeader, types.String(reader.ExpiresAt()))
	}

	// 输出Body
	if this.RawReq.Method == http.MethodHead {
		this.writer.WriteHeader(reader.Status())
	} else {
		ifRangeHeaders, ok := this.RawReq.Header["If-Range"]
		var supportRange = true
		if ok {
			supportRange = false
			for _, v := range ifRangeHeaders {
				if v == this.writer.Header().Get("ETag") || v == this.writer.Header().Get("Last-Modified") {
					supportRange = true
					break
				}
			}
		}

		// 支持Range
		var ranges = partialRanges
		if supportRange {
			if len(rangeHeader) > 0 {
				if fileSize == 0 {
					this.ProcessResponseHeaders(this.writer.Header(), http.StatusRequestedRangeNotSatisfiable)
					this.writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return true
				}

				if len(ranges) == 0 {
					ranges, ok = httpRequestParseRangeHeader(rangeHeader)
					if !ok {
						this.ProcessResponseHeaders(this.writer.Header(), http.StatusRequestedRangeNotSatisfiable)
						this.writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
						return true
					}
				}
				if len(ranges) > 0 {
					for k, r := range ranges {
						r2, ok := r.Convert(fileSize)
						if !ok {
							this.ProcessResponseHeaders(this.writer.Header(), http.StatusRequestedRangeNotSatisfiable)
							this.writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
							return true
						}

						ranges[k] = r2
					}
				}
			}
		}

		if len(ranges) == 1 {
			respHeader.Set("Content-Range", ranges[0].ComposeContentRangeHeader(totalSizeString))
			respHeader.Set("Content-Length", strconv.FormatInt(ranges[0].Length(), 10))
			this.writer.WriteHeader(http.StatusPartialContent)

			var pool = this.bytePool(fileSize)
			var bodyBuf = pool.Get()

			var rangeEnd = ranges[0].End()
			if firstRangeEnd > 0 {
				rangeEnd = firstRangeEnd
			}

			err = reader.ReadBodyRange(bodyBuf.Bytes, ranges[0].Start(), rangeEnd, func(n int) (goNext bool, readErr error) {
				_, readErr = this.writer.Write(bodyBuf.Bytes[:n])
				if readErr != nil {
					return false, errWritingToClient
				}
				return true, nil
			})
			pool.Put(bodyBuf)
			if err != nil {
				this.varMapping["cache.status"] = "MISS"

				if errors.Is(err, caches.ErrInvalidRange) {
					this.ProcessResponseHeaders(this.writer.Header(), http.StatusRequestedRangeNotSatisfiable)
					this.writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return true
				}
				if !this.canIgnore(err) {
					remotelogs.WarnServer("HTTP_REQUEST_CACHE", this.URL()+": read from cache failed: "+err.Error())
				}

				return true
			}
		} else if len(ranges) > 1 {
			var boundary = httpRequestGenBoundary()
			respHeader.Set("Content-Type", "multipart/byteranges; boundary="+boundary)
			respHeader.Del("Content-Length")
			var contentType = respHeader.Get("Content-Type")

			this.writer.WriteHeader(http.StatusPartialContent)

			for index, r := range ranges {
				if index == 0 {
					_, err = this.writer.WriteString("--" + boundary + "\r\n")
				} else {
					_, err = this.writer.WriteString("\r\n--" + boundary + "\r\n")
				}
				if err != nil {
					// 不提示写入客户端错误
					return true
				}

				_, err = this.writer.WriteString("Content-Range: " + r.ComposeContentRangeHeader(totalSizeString) + "\r\n")
				if err != nil {
					// 不提示写入客户端错误
					return true
				}

				if len(contentType) > 0 {
					_, err = this.writer.WriteString("Content-Type: " + contentType + "\r\n\r\n")
					if err != nil {
						// 不提示写入客户端错误
						return true
					}
				}

				var pool = this.bytePool(fileSize)
				var bodyBuf = pool.Get()
				err = reader.ReadBodyRange(bodyBuf.Bytes, r.Start(), r.End(), func(n int) (goNext bool, readErr error) {
					_, readErr = this.writer.Write(bodyBuf.Bytes[:n])
					if readErr != nil {
						return false, errWritingToClient
					}
					return true, nil
				})
				pool.Put(bodyBuf)
				if err != nil {
					if !this.canIgnore(err) {
						remotelogs.WarnServer("HTTP_REQUEST_CACHE", this.URL()+": read from cache failed: "+err.Error())
					}
					return true
				}
			}

			_, err = this.writer.WriteString("\r\n--" + boundary + "--\r\n")
			if err != nil {
				this.varMapping["cache.status"] = "MISS"

				// 不提示写入客户端错误
				return true
			}
		} else { // 没有Range
			var resp = &http.Response{
				Body:          reader,
				ContentLength: reader.BodySize(),
			}
			this.writer.Prepare(resp, fileSize, reader.Status(), false)
			this.writer.WriteHeader(reader.Status())

			if storage.CanSendfile() {
				var pool = this.bytePool(fileSize)
				var bodyBuf = pool.Get()
				if fp, canSendFile := this.writer.canSendfile(); canSendFile {
					this.writer.sentBodyBytes, err = io.CopyBuffer(this.writer.rawWriter, fp, bodyBuf.Bytes)
				} else {
					_, err = io.CopyBuffer(this.writer, resp.Body, bodyBuf.Bytes)
				}
				pool.Put(bodyBuf)
			} else {
				mmapReader, isMMAPReader := reader.(*caches.MMAPFileReader)
				if isMMAPReader {
					_, err = mmapReader.CopyBodyTo(this.writer)
				} else {
					var pool = this.bytePool(fileSize)
					var bodyBuf = pool.Get()
					_, err = io.CopyBuffer(this.writer, resp.Body, bodyBuf.Bytes)
					pool.Put(bodyBuf)
				}
			}

			if err == io.EOF {
				err = nil
			}
			if err != nil {
				this.varMapping["cache.status"] = "MISS"

				if !this.canIgnore(err) {
					remotelogs.WarnServer("HTTP_REQUEST_CACHE", this.URL()+": read from cache failed: read body failed: "+err.Error())
				}
				return
			}
		}
	}

	this.isCached = true
	this.cacheRef = nil

	this.writer.SetOk()

	return true
}

// 设置Expires Header
func (this *HTTPRequest) addExpiresHeader(expiresAt int64) {
	if this.cacheRef.ExpiresTime != nil && this.cacheRef.ExpiresTime.IsPrior && this.cacheRef.ExpiresTime.IsOn {
		if this.cacheRef.ExpiresTime.Overwrite || len(this.writer.Header().Get("Expires")) == 0 {
			if this.cacheRef.ExpiresTime.AutoCalculate {
				this.writer.Header().Set("Expires", time.Unix(utils.GMTUnixTime(expiresAt), 0).Format("Mon, 2 Jan 2006 15:04:05")+" GMT")
				this.writer.Header().Del("Cache-Control")
			} else if this.cacheRef.ExpiresTime.Duration != nil {
				var duration = this.cacheRef.ExpiresTime.Duration.Duration()
				if duration > 0 {
					this.writer.Header().Set("Expires", utils.GMTTime(time.Now().Add(duration)).Format("Mon, 2 Jan 2006 15:04:05")+" GMT")
					this.writer.Header().Del("Cache-Control")
				}
			}
		}
	}
}

// 尝试读取区间缓存
func (this *HTTPRequest) tryPartialReader(storage caches.StorageInterface, key string, useStale bool, rangeHeader string, forcePartialContent bool) (resultReader caches.Reader, ranges []rangeutils.Range, firstRangeEnd int64, goNext bool) {
	goNext = true

	// 尝试读取Partial cache
	if len(rangeHeader) == 0 {
		return
	}

	ranges, ok := httpRequestParseRangeHeader(rangeHeader)
	if !ok {
		return
	}

	pReader, pErr := storage.OpenReader(key+caches.SuffixPartial, useStale, true)
	if pErr != nil {
		if caches.IsBusyError(pErr) {
			this.varMapping["cache.status"] = "BUSY"
			goNext = false
			return
		}
		return
	}

	partialReader, ok := pReader.(*caches.PartialFileReader)
	if !ok {
		_ = pReader.Close()
		return
	}
	var isOk = false
	defer func() {
		if !isOk {
			_ = pReader.Close()
		}
	}()

	// 检查是否已下载完整
	if !forcePartialContent &&
		len(ranges) > 0 &&
		ranges[0][1] < 0 &&
		!partialReader.IsCompleted() {
		if partialReader.BodySize() > 0 {
			var options = this.ReqServer.HTTPCachePolicy.Options
			if options != nil {
				fileStorage, isFileStorage := storage.(*caches.FileStorage)
				if isFileStorage && fileStorage.Options() != nil && fileStorage.Options().EnableIncompletePartialContent {
					var r = ranges[0]
					r2, findOk := partialReader.Ranges().FindRangeAtPosition(r.Start())
					if findOk && r2.Length() >= (256<<10) /* worth reading */ {
						isOk = true
						ranges[0] = [2]int64{r.Start(), partialReader.BodySize() - 1} // Content-Range: bytes 0-[CONTENT_LENGTH - 1]/CONTENT_LENGTH

						pReader.SetNextReader(NewHTTPRequestPartialReader(this, r2.End(), partialReader))
						return pReader, ranges, r2.End() - 1 /* not include last byte */, true
					}
				}
			}
		}
		return
	}

	// 检查范围
	// 这里 **切记不要** 为末尾位置指定一个中间值，因为部分软件客户端不支持
	for index, r := range ranges {
		r1, ok := r.Convert(partialReader.MaxLength())
		if !ok {
			return
		}
		r2, ok := partialReader.ContainsRange(r1)
		if !ok {
			return
		}
		ranges[index] = r2
	}

	isOk = true
	return pReader, ranges, -1, true
}
