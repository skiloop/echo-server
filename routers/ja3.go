package routers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/skiloop/echo-server/server"
	"net/http"
)

func JA3(c echo.Context) error {
	j := c.Get("ja3")
	if j == nil {
		return c.JSON(http.StatusInternalServerError, "{\"code\":100,\"message\":\"no ja3\"}")
	}
	ja3 := j.(server.Ja3)
	c.Logger().Debugf("ja3 %s", ja3.Md5Hash())
	return c.JSON(http.StatusOK, ja3)
}

func SetJA3Routers(c *echo.Echo, prefix string) {
	if prefix == "/" {
		prefix = ""
	}
	c.GET(fmt.Sprintf("%s/ja3", prefix), JA3)
}
