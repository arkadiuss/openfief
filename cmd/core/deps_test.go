package main

import (
	"reflect"
	"testing"
)

// TestCoreConnectorDepsComplete starts the core and checks that every
// dependency it hands its connector scheduler is set: a nil one only
// fails when a scheduled sync reaches it, as the tax wrappers did, which
// the core never set while the connector command did (D-117).
func TestCoreConnectorDepsComplete(t *testing.T) {
	t.Parallel()
	c, _, _ := startCoreForUI(t, emptyDatabase(t), func(*serveConfig) {})
	v := reflect.ValueOf(c.deps).Elem()
	for i := range v.NumField() {
		if v.Field(i).IsZero() {
			t.Errorf("the core's connector dependency %s is not set", v.Type().Field(i).Name)
		}
	}
}
