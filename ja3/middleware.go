package ja3

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

const XJa3HeaderName = "X-JA3"
const XJa3HashKey = "X-JA3-HASH"
const XJa3Key = "X-JA3"

func JA3(c echo.Context) error {
	j := c.Get(XJa3Key)
	if j == nil {
		return c.JSON(http.StatusOK, "{\"code\":100,\"message\":\"no ja3\"}")
	}
	j3 := j.(Ja3)
	c.Logger().Debugf("ja3 %s", j3.Md5Hash())
	return c.JSON(http.StatusOK, j3)
}

func SetJA3Routers(c *echo.Echo, prefix string) {
	if prefix == "/" {
		prefix = ""
	}
	c.GET(fmt.Sprintf("%s/ja3", prefix), JA3)
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			k := context.Request().RemoteAddr
			//store := context.Get("store").(echo.Map)
			// TODO: conflicts appear when different client from the same remote address
			if val, ok := store.Get(k); ok {
				context.Set(XJa3Key, val)

				val := val.(Ja3)
				valHash := val.Md5Hash()
				context.Set(XJa3HashKey, valHash)

				context.Echo().Logger.Debugf("get %s ja3 for %s", k, valHash)
				context.Response().Header().Set(XJa3HeaderName, valHash)
			} else {
				context.Echo().Logger.Warnf("no ja3 for %s", k)
			}
			return next(context)
		}
	}
}
