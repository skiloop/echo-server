package routers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/skiloop/echo-server/version"
)

// GetVersion
// get version
func GetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"version":     version.Version,
		"commit_hash": version.CommitHash,
		"build_time":  version.BuildTime,
	})
}
