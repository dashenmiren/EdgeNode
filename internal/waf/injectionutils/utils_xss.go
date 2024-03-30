package injectionutils

/*
#cgo CFLAGS: -O2 -I./libinjection/src

#include <libinjection.h>
#include <stdlib.h>
*/
import "C"
import (
	"net/url"
	"strconv"
	"strings"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	"github.com/dashenmiren/EdgeNode/internal/waf/utils"
)

func DetectXSSCache(input string, isStrict bool, cacheLife utils.CacheLife) bool {
	var l = len(input)

	if l == 0 {
		return false
	}

	if cacheLife <= 0 || l < 512 || l > utils.MaxCacheDataSize {
		return DetectXSS(input, isStrict)
	}

	var hash = xxhash.Sum64String(input)
	var key = "WAF@XSS@" + strconv.FormatUint(hash, 10)
	if isStrict {
		key += "@1"
	}
	var item = utils.SharedCache.Read(key)
	if item != nil {
		return item.Value == 1
	}

	var result = DetectXSS(input, isStrict)
	if result {
		utils.SharedCache.Write(key, 1, fasttime.Now().Unix()+cacheLife)
	} else {
		utils.SharedCache.Write(key, 0, fasttime.Now().Unix()+cacheLife)
	}
	return result
}

// DetectXSS detect XSS in string
func DetectXSS(input string, isStrict bool) bool {
	if len(input) == 0 {
		return false
	}

	if detectXSSOne(input, isStrict) {
		return true
	}

	// 兼容 /PATH?URI
	if (input[0] == '/' || strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")) && len(input) < 1024 {
		var argsIndex = strings.Index(input, "?")
		if argsIndex > 0 {
			var args = input[argsIndex+1:]
			unescapeArgs, err := url.QueryUnescape(args)
			if err == nil && args != unescapeArgs {
				return detectXSSOne(args, isStrict) || detectXSSOne(unescapeArgs, isStrict)
			} else {
				return detectXSSOne(args, isStrict)
			}
		}
	} else {
		unescapedInput, err := url.QueryUnescape(input)
		if err == nil && input != unescapedInput {
			return detectXSSOne(unescapedInput, isStrict)
		}
	}

	return false
}

func detectXSSOne(input string, isStrict bool) bool {
	if len(input) == 0 {
		return false
	}

	var cInput = C.CString(input)
	defer C.free(unsafe.Pointer(cInput))

	var isStrictInt = 0
	if isStrict {
		isStrictInt = 1
	}
	return C.libinjection_xss(cInput, C.size_t(len(input)), C.int(isStrictInt)) == 1
}
