package routers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin} // use default options
const wsMessageLimit = 100

func checkOrigin(r *http.Request) bool{
	return true
}

// WsEcho
// websocket echo
func WsEcho(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Warn("upgrade error:", err)
		return err
	}
	defer func(con *websocket.Conn) {
		err := con.Close()
		if err != nil {
			c.Logger().Error("websocket close error:", err)
		}
	}(conn)
	err = writeCommonMessage(conn)
	if err != nil {
		c.Logger().Error("common message write error:", err)
		return err
	}
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			c.Logger().Error("read:", err)
			break
		}
		c.Logger().Debugf("recv: type %d, size %d, message: %s", mt, len(message), message)
		if len(message) > wsMessageLimit {
			message = message[:wsMessageLimit]
		}
		err = conn.WriteMessage(mt, message)
		if err != nil {
			c.Logger().Error("write:", err)
			break
		}
	}
	return nil
}

func writeCommonMessage(conn *websocket.Conn) error {
	message := []byte(fmt.Sprintf("message limit: %d", wsMessageLimit))
	return conn.WriteMessage(websocket.TextMessage, message)
}
