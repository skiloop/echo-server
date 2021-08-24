package routers

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/maxmind/mmdbinspect/pkg/mmdbinspect"
	"github.com/skiloop/echo-server/utils"
	"net/http"
	"os"
	"strings"
)

var geoLite2Path = "./"
var EnableGeoIP = false

func init() {
	geoPath := os.Getenv("GEO_LITE_2_PATH")
	if geoPath != "" {
		geoLite2Path = geoPath
	}
	EnableGeoIP = utils.Exists(geoPath)
}

type City struct {
	GeoNameID int               `json:"geoname_id"`
	Names     map[string]string `json:"names"`
}
type Continent struct {
	Code      string            `json:"code"`
	GeoNameID int               `json:"geoname_id"`
	Names     map[string]string `json:"names"`
}

type Place struct {
	IsoCode   string            `json:"iso_code"`
	GeoNameID int               `json:"geoname_id"`
	Names     map[string]string `json:"names"`
}

type Coordinate struct {
	AccuracyRadius int     `json:"accuracy_radius"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	TimeZone       string  `json:"time_zone"`
}

type Location struct {
	City              City       `json:"city"`
	Continent         Continent  `json:"continent"`
	Country           Place      `json:"country"`
	Location          Coordinate `json:"location"`
	RegisteredCountry Place      `json:"registered_country"`
	Subdivisions      []Place    `json:"subdivisions"`
}

type Record struct {
	Network string   `json:"Network"`
	Record  Location `json:"Record"`
}

type ipResponse struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Network   string  `json:"network"`
	IP        string  `json:"ip"`
	YourIP    string  `json:"your_ip"`
}

func getWithDefaultLang(names map[string]string, lang, defaultLang string) string {
	if names == nil {
		return ""
	}
	if _, has := names[lang]; has {
		return names[lang]
	} else if _, has = names[defaultLang]; has {
		return names[defaultLang]
	}
	return ""
}

func LookUp(ip string) (*Record, error) {
	dbPath := geoLite2Path + "/" + "GeoLite2-City.mmdb"
	networks := make([]string, 1)
	dbPaths := make([]string, 1)
	dbPaths[0] = dbPath
	networks[0] = ip
	resSet, err := mmdbinspect.AggregatedRecords(networks, dbPaths)
	if err != nil {
		return nil, err
	}
	records := resSet.([]mmdbinspect.RecordSet)
	if 0 == len(records) {
		return nil, errors.New("no data")
	}
	rs := records[0].Records.([]interface{})
	if 0 == len(rs) {
		return nil, errors.New("no data")
	}
	log.Debug(rs[0])
	rc, err := json.Marshal(rs[0])
	if err != nil {
		return nil, err
	}
	var record Record
	err = json.Unmarshal(rc, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func formResponse(lang, yourIP, ip string, record *Record) *ipResponse {
	var r ipResponse
	en := "en"
	r.Country = getWithDefaultLang(record.Record.Country.Names, lang, en)
	r.City = getWithDefaultLang(record.Record.City.Names, lang, en)
	r.Latitude = record.Record.Location.Latitude
	r.Longitude = record.Record.Location.Longitude
	r.Network = record.Network
	r.YourIP = yourIP
	r.IP = ip
	return &r
}
func noQueryResponse(message, realIp string) map[string]interface{} {
	m := make(map[string]interface{})
	m["code"] = 1
	m["message"] = message
	m["your_id"] = realIp
	return m
}

func GetIp(c echo.Context) (err error) {
	realIP := c.RealIP()
	c.Logger().Debugf("real ip: %s", realIP)
	var src string
	// path parameter
	src = c.Param("ip")
	if src == "" {
		// query parameter
		src = c.QueryParam("ip")
	}
	if src == "" {
		src = realIP
	}
	if src == "" {
		_ = c.JSON(http.StatusOK, noQueryResponse("ip error", realIP))
		return
	}
	if !EnableGeoIP {
		_ = c.JSON(http.StatusOK, noQueryResponse("geoip disabled", realIP))
		return
	}

	record, err := LookUp(src)
	if err != nil {
		_ = c.JSON(http.StatusOK, noQueryResponse("no data", realIP))
		return
	}
	// lang
	var lang string
	lang = c.QueryParam("lang")
	if lang == "" {
		lang = c.Request().Header.Get("Accept-Language")
		if lang != "" {
			lang = strings.SplitN(lang, ",", 2)[0]
		}
	}
	if lang == "" {
		lang = "en"
	}
	log.Infof("lang: %s", lang)
	_ = c.JSON(http.StatusOK, formResponse(lang, realIP, src, record))
	return
}

func YourIp(c echo.Context) error {
	return c.String(http.StatusOK, c.RealIP())
}
