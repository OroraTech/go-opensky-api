package opensky

import (
	"strconv"
	"time"
)

// Utility time wrapper struct, needed for unmarshaling unix
// timestamps from JSON objects.
type UnixTime struct {
	time.Time
}

// Helper function for generating a UnixTime struct from a unix timestamp.
func newUnixTime(sec int64) UnixTime {
	return UnixTime{time.Unix(sec, 0)}
}

// Helper function for generating a pointer to UnixTime struct from a unix timestamp.
func newUnixTimeP(sec int64) *UnixTime {
	return &UnixTime{time.Unix(sec, 0)}
}

func (t *UnixTime) UnmarshalJSON(s []byte) error {
	raw := string(s)
	if raw == "null" {
		*t = UnixTime{time.Time{}}
		return nil
	}
	unixTimestamp, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime{time.Unix(unixTimestamp, 0)}
	return nil
}
