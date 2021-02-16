package time

import (
	"context"
	"database/sql/driver"
	"strconv"
	"time"
)

// Time be used to mysql timestamp converting.
type Time int64

// Scan scan time.
func (t *Time) Scan(src interface{}) error {
	switch sc := src.(type) {
	case time.Time:
		*t = Time(sc.Unix())
	case string:
		i, err := strconv.ParseInt(sc, 10, 64)
		if err != nil {
			return err
		}
		*t = Time(i)
	}
	return nil
}

// Value get time value.
func (t Time) Value() (driver.Value, error) {
	return time.Unix(int64(t), 0), nil
}

// Time get time.
func (t Time) Time() time.Time {
	return time.Unix(int64(t), 0)
}

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration time.Duration

// Duration returns a time.Duration.
func (d *Duration) Duration() time.Duration {
	return time.Duration(*d)
}

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(tmp)
	return nil
}

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if timeout := time.Until(deadline); timeout < time.Duration(d) {
			return Duration(timeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}
