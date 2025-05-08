package ja3

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

const XJa3HeaderName = "X-JA3"

func JA3(c echo.Context) error {
	j := c.Get("ja3")
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

func HashMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			k := context.Request().RemoteAddr
			//store := context.Get("store").(echo.Map)
			// TODO: conflicts appear when different client from the same remote address
			if val, ok := store[k]; ok {
				delete(store, k)
				context.Set("ja3", val)
				val := val.(Ja3)
				context.Echo().Logger.Debugf("get %s ja3 for %s", k, val.Md5Hash())
				context.Response().Header().Set(XJa3HeaderName, val.Md5Hash())
			} else {
				context.Echo().Logger.Warnf("no ja3 for %s", k)
			}
			return next(context)
		}
	}
}
