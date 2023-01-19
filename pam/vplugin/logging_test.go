package vplugin

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestZerologAdapter(t *testing.T) {
	captor := strings.Builder{}
	log.Logger = log.Logger.Level(zerolog.TraceLevel).Output(&captor)
	defer func() {
		// switch back to INFO after test, so we don't slow others down
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}()
	adapter := NewZerologAdapter()

	tests := []struct {
		name  string
		logFn func()
		want  string
	}{
		{
			"log level",
			func() { adapter.Log(hclog.Debug, "level") },
			"\"level\":\"debug\"",
		},
		{
			"log trace",
			func() { adapter.Trace("tracemsg") },
			"tracemsg",
		},
		{
			"log debug",
			func() { adapter.Debug("debugmsg") },
			"debugmsg",
		},
		{
			"log info",
			func() { adapter.Info("infomsg") },
			"infomsg",
		},
		{
			"log warn",
			func() { adapter.Warn("warnmsg") },
			"warnmsg",
		},
		{
			"log error",
			func() { adapter.Error("errormsg") },
			"errormsg",
		},
		{
			"log args",
			func() { adapter.Info("msg", "key", "value") },
			"\"key\":\"value\"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			captor.Reset()
			test.logFn()
			assert.Contains(t, captor.String(), test.want)
		})
	}
}

func TestZerologAdapter_SkippedTimeArgs(t *testing.T) {
	captor := strings.Builder{}
	log.Logger = log.Logger.Output(&captor)
	adapter := NewZerologAdapter()

	adapter.Info("msg", "time", "value")
	assert.NotContains(t, captor.String(), "\"time\":\"value\"")
}

func TestZerologAdapter_SkippedTimestampArgs(t *testing.T) {
	captor := strings.Builder{}
	log.Logger = log.Logger.Output(&captor)
	adapter := NewZerologAdapter()

	adapter.Info("msg", "timestamp", "value")
	assert.NotContains(t, captor.String(), "\"timestamp\":\"value\"")
}

func TestZerologAdapter_Level(t *testing.T) {
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
	adapter := NewZerologAdapter()

	assert.False(t, adapter.IsTrace())
	assert.False(t, adapter.IsDebug())
	assert.True(t, adapter.IsInfo())
	assert.True(t, adapter.IsWarn())
	assert.True(t, adapter.IsError())
	assert.Equal(t, hclog.Info, adapter.GetLevel())
}
