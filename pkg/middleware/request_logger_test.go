package middleware

import (
	"bytes"
	"testing"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestLogger_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableHandler(RequestLogger)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_RequestLogger_PrintsRequestTiming(t *testing.T) {
	callable, _, ctx := createCallableHandler(RequestLogger)

	var out bytes.Buffer
	log := logger.New(&out)
	ctx.SetLogger(logger.Wrap(log))

	err := callable(ctx)
	require.Nil(t, err)
	afterCall := time.Now()

	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "info", actual.Level)
	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
	assert.Regexp(t, `GET example.com/ processed in [0-9]+(\.[0-9]+)?([mµn])?s -> \x1b\[1;32m200\x1b\[0m`, actual.Message)
}

func areTimeCloserThan(t1 time.Time, t2 time.Time, distance time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= distance
}
