package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/skiloop/echo-server/ja3"
	"github.com/skiloop/echo-server/utils"
	"sync/atomic"
)

var wafStore *utils.KVStore
var requestKeyGenerator func(c echo.Context) string

var wafReqIdKey = "waf-req-id"

type meta struct {
	ip    string
	count *atomic.Uint32
}

func middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		rid := context.Get(wafReqIdKey)
		var reqId string
		var ok bool

		if rid == nil {
			if requestKeyGenerator != nil {
				reqId = requestKeyGenerator(context)
			} else {
				context.Logger().Debugf("no available request id use remote address instead")
				reqId = context.Request().RemoteAddr
			}
		} else {
			reqId, ok = rid.(string)
			if !ok {
				context.Logger().Warnf("failed to convert request id for %s", context.Request().RemoteAddr)
				return next(context)
			}
		}
		m, ok := wafStore.Get(reqId)
		if !ok {
			m = &meta{ip: context.Request().RemoteAddr, count: &atomic.Uint32{}}
			wafStore.Set(reqId, m)
		}
		mt := m.(*meta)
		mt.count.Add(1)
		context.Logger().Debugf("[%s] count %d", reqId, mt.count.Load())
		return next(context)
	}
}

// WafMiddleware is a middleware that provides Web Application Firewall (WAF) functionality.
func WafMiddleware(ttl int64, keyFunc func(ctx echo.Context) string, ctx context.Context) echo.MiddlewareFunc {
	if wafStore == nil {
		wafStore = utils.NewKVStore(ttl)
		wafStore.StartCleanupRoutine(ctx, 600)
		if keyFunc == nil {
			wafReqIdKey = ja3.XJa3HashKey
		} else {
			requestKeyGenerator = keyFunc
		}
	}
	return middleware
}
