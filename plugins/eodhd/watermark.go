package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// errBadWatermark reports a watermark this connector never returned.
var errBadWatermark = errors.New("watermark is not this connector's")

// watermark is the connector's watermark: the date of the latest close
// delivered for each instrument, keyed by ISIN, and for each index, keyed
// by its name. It travels as JSON; encoding/json writes map keys sorted, so
// the same marks always give the same text.
type watermark struct {
	Prices  map[string]string `json:"prices,omitempty"`
	Indices map[string]string `json:"indices,omitempty"`
}

// parseWatermark reads a watermark; empty is a first sync. Every mark must
// be a date.
func parseWatermark(s string) (watermark, error) {
	w := watermark{Prices: map[string]string{}, Indices: map[string]string{}}
	if s == "" {
		return w, nil
	}
	if err := json.Unmarshal([]byte(s), &w); err != nil {
		return watermark{}, fmt.Errorf("%w: %w", errBadWatermark, err)
	}
	if w.Prices == nil {
		w.Prices = map[string]string{}
	}
	if w.Indices == nil {
		w.Indices = map[string]string{}
	}
	for _, marks := range []map[string]string{w.Prices, w.Indices} {
		for key, mark := range marks {
			if _, err := kernel.ParseDate(mark); err != nil {
				return watermark{}, fmt.Errorf("%w: %s: %w", errBadWatermark, key, err)
			}
		}
	}
	return w, nil
}

// nextFrom is the first day to ask for key, whose marks are in marks: the
// day after its mark, or start without one; and whether that day is due,
// on or before today.
func nextFrom(marks map[string]string, key string, start, today kernel.Date) (kernel.Date, bool) {
	mark, ok := marks[key]
	if !ok {
		return start, !start.After(today)
	}
	last, err := kernel.ParseDate(mark)
	if err != nil { // parseWatermark checked every mark
		return start, !start.After(today)
	}
	next := last.AddDays(1)
	return next, !next.After(today)
}

// String is the watermark's text; no marks at all is empty, like a first
// sync's.
func (w watermark) String() string {
	if len(w.Prices) == 0 && len(w.Indices) == 0 {
		return ""
	}
	b, err := json.Marshal(w)
	if err != nil { // maps of strings always marshal
		return ""
	}
	return string(b)
}
