package caleta

import (
	"strings"
	"time"
)

// CaletaDate is a type alias for time.Time and only used to trick OAPI model
// generation from referencing time.Time directly. Thus not allowing custom
// unmarshaller.
type caletaDate time.Time

const format = "2006-01-02 15:04:05.999999-07:00"

func (c *MsgTimestamp) UnmarshalJSON(in []byte) error {
	s := strings.Trim(string(in), "\"")
	tt, err := time.Parse(format, s)
	if err != nil {
		return err
	}
	*c = MsgTimestamp(tt)
	return nil
}

func (c MsgTimestamp) toTime() time.Time {
	return time.Time(c)
}
