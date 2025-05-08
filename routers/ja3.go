package routers

import (
	"fmt"
	"github.com/skiloop/echo-server/ja3"
	"net/http"
)

func JA3(c echo.Context) error {
	j := c.Get("ja3")
	if j == nil {
		return c.JSON(http.StatusOK, "{\"code\":100,\"message\":\"no ja3\"}")
	}
	j3 := j.(ja3.Ja3)
	c.Logger().Debugf("ja3 %s", j3.Md5Hash())
	return c.JSON(http.StatusOK, j3)
}

func SetJA3Routers(c *echo.Echo, prefix string) {
	if prefix == "/" {
		prefix = ""
	}
	c.GET(fmt.Sprintf("%s/ja3", prefix), JA3)
}
