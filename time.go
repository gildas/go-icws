package icws

import (
	"strings"
	"time"

	"github.com/gildas/go-errors"
)

type Time time.Time

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format("\"20060102T150405Z\"")), nil
}

// UnmarshalJSON unmarshals from JSON
//
// implements json.Unmarshaler
func (t *Time) UnmarshalJSON(payload []byte) (err error) {
	tt, err := time.Parse("20060102T150405Z", strings.Trim(string(payload), "\""))
	if err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*t = Time(tt)
	return nil
}
