package approvalrulesengine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseScheduleTimeTooFewComponents(t *testing.T) {
	inputs := []string{"", "1"}
	for _, input := range inputs {
		_, err := parseScheduleTime(time.Now(), input)
		if assert.NotNil(t, err, "Input=%s", input) {
			assert.Regexp(t, "Invalid time format", err.Error(), "Input=%s", input)
		}
	}
}

func TestParseScheduleTimeInvalidValues(t *testing.T) {
	type Input struct {
		Value     string
		Component string
	}

	inputs := []Input{
		{Value: ":", Component: "hour"},
		{Value: "a:", Component: "hour"},
		{Value: "-1:", Component: "hour"},
		{Value: "25:", Component: "hour"},

		{Value: "1:", Component: "minute"},
		{Value: "1:b", Component: "minute"},
		{Value: "1:-1", Component: "minute"},
		{Value: "1:61", Component: "minute"},

		{Value: "1:30:", Component: "second"},
		{Value: "1:30:c", Component: "second"},
		{Value: "1:30:-1", Component: "second"},
		{Value: "1:30:61", Component: "second"},
	}

	for _, input := range inputs {
		_, err := parseScheduleTime(time.Now(), input.Value)
		if assert.NotNil(t, err, "Input=%#v", input) {
			assert.Regexp(t, "Error parsing "+input.Component+" component",
				err.Error(), "Input=%#v", input)
		}
	}
}

func TestParseScheduleTimeValidValues(t *testing.T) {
	var err error
	var parsed time.Time
	now := time.Now()

	parsed, err = parseScheduleTime(now, "1:20")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 1)
		assert.Equal(t, parsed.Minute(), 20)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "01:20")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 1)
		assert.Equal(t, parsed.Minute(), 20)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "16:5")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 16)
		assert.Equal(t, parsed.Minute(), 5)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "16:05")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 16)
		assert.Equal(t, parsed.Minute(), 5)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "8:47:1")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 8)
		assert.Equal(t, parsed.Minute(), 47)
		assert.Equal(t, parsed.Second(), 1)
	}

	parsed, err = parseScheduleTime(now, "8:47:01")
	if assert.Nil(t, err) {
		assert.Equal(t, parsed.Hour(), 8)
		assert.Equal(t, parsed.Minute(), 47)
		assert.Equal(t, parsed.Second(), 1)
	}
}
