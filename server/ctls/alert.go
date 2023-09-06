package ctls

import "strconv"

type alert uint8

const (
	alertUnexpectedMessage alert = 10
	alertBadRecordMAC      alert = 20
	alertInternalError     alert = 80
)

var alertText = map[alert]string{
	alertUnexpectedMessage: "unexpected message",
	alertBadRecordMAC:      "bad record MAC",
	alertInternalError:     "internal error",
}

func (e alert) String() string {
	s, ok := alertText[e]
	if ok {
		return "tls: " + s
	}
	return "tls: alert(" + strconv.Itoa(int(e)) + ")"
}

func (e alert) Error() string {
	return e.String()
}
