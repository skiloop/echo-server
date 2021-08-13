package routers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func MongoParseID(c echo.Context) error {
	objId := c.Param("id")
	if len(objId) != 24 {
		return c.JSON(http.StatusBadRequest, "{'code':1,'message':'invalid id'}")
	}
	obj, err := ParseMongoObjectID(objId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "{'code':1,'message':'invalid id string'}")
	}
	c.Logger().Debugf("parse mongodb object id: %s", objId)
	return c.JSON(http.StatusOK, obj)
}

type ObjectId struct {
	Timestamp int64 `json:"timestamp"`
	Host      int32 `json:"host"`
	PID       int32 `json:"pid"`
	RandomInt int32 `json:"random_int"`
}

func ParseMongoObjectID(id string) (*ObjectId, error) {

	obj := ObjectId{}
	var err error
	var t int64
	obj.Timestamp, err = strconv.ParseInt(id[:8], 16, 64)
	if err != nil {
		return nil, err
	}

	t, err = strconv.ParseInt(id[8:14], 16, 32)
	if err != nil {
		return nil, err
	}
	obj.Host = int32(t)

	t, err = strconv.ParseInt(id[14:18], 16, 32)
	if err != nil {
		return nil, err
	}
	obj.PID = int32(t)
	t, err = strconv.ParseInt(id[18:], 16, 32)
	if err != nil {
		return nil, err
	}
	obj.RandomInt = int32(t)
	return &obj, nil
}
