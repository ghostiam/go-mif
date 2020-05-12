package mif

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestMIFWithoutOptions(t *testing.T) {
	m := New().(*mif)

	equals(t, noopLogger{}, m.logger)
	equals(t, JSONConfig{}, m.json)
	assert(t, !m.disablePanic, "disablePanic should be false")
}

func TestWithLogger(t *testing.T) {
	l := &testLogger{}
	m := New(WithLogger(l)).(*mif)

	equals(t, l, m.logger)
}

func TestWithStdLogger(t *testing.T) {
	l := log.New(ioutil.Discard, "", 0)
	m := New(WithStdLogger(l)).(*mif)

	equals(t, l, m.logger.(stdLoggerWrapper).logger)
}

func TestWithNoopLogger(t *testing.T) {
	m := New(WithNoopLogger()).(*mif)

	equals(t, noopLogger{}, m.logger)
}

func TestWithJSONConfig(t *testing.T) {
	cfg := JSONConfig{
		Prefix: "prefix",
		Indent: "\t",
	}
	m := New(WithJSONConfig(cfg)).(*mif)

	equals(t, cfg, m.json)
}

func TestSetDisablePanic(t *testing.T) {
	m := New(SetDisablePanic()).(*mif)

	assert(t, m.disablePanic, "SetDisablePanic should disable panic")
}
