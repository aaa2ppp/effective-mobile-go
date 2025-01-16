package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"effective-mobile-go/internal/lib"
)

type Date struct {
	time.Time `json:"-"`
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse("02.01.2006", s)
	return Date{t}, err
}

func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d.%02d.%04d", d.Day(), d.Month(), d.Year())
}

// MarshalJSON implements json.Marshaler.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	w := bytes.NewBuffer(make([]byte, 0, 10))
	fmt.Fprintf(w, `"%02d.%02d.%04d"`, d.Day(), d.Month(), d.Year())
	return w.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := lib.UnsafeString(b)
	if s == "null" {
		return nil // no-op
	}
	s, err := strconv.Unquote(s)
	if err != nil {
		return err
	}
	v, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = v
	return nil
}

var (
	_ json.Marshaler   = Date{}
	_ json.Unmarshaler = (*Date)(nil)
)
