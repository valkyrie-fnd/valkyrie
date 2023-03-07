package ops

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

var trueVar = true
var falseVar = false

func Test_getFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name   string
		config configs.OutputLogConfig
		want   io.Writer
	}{
		{
			"success with defaults",
			configs.OutputLogConfig{
				Filename: tempDir + string(os.PathSeparator) + "valkyrie.log",
			},
			&lumberjack.Logger{
				Filename: tempDir + string(os.PathSeparator) + "valkyrie.log",
			},
		},
		{
			"success with all configured",
			configs.OutputLogConfig{
				Type:       "file",
				Filename:   tempDir + string(os.PathSeparator) + "test",
				MaxSize:    1,
				MaxAge:     2,
				MaxBackups: 3,
				Compress:   true,
			},
			&lumberjack.Logger{
				Filename:   tempDir + string(os.PathSeparator) + "test",
				MaxSize:    1,
				MaxAge:     2,
				MaxBackups: 3,
				Compress:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getFileWriter(tt.config), "getFileWriter(%v)", tt.config)
		})
	}
}

func TestConfigureLogging(t *testing.T) {
	// Configure logger with fixed timestamp generator function
	expectedTimestampString := "2017-11-15T13:16:30.646Z"
	expectedTimestamp, _ := time.Parse(time.RFC3339Nano, expectedTimestampString)
	zerolog.TimestampFunc = func() time.Time {
		return expectedTimestamp
	}

	// Reset global logger to its original state when done testing
	globalLogger := log.Logger
	defer func() {
		log.Logger = globalLogger
	}()

	// Create temporary file to log to
	file, err := os.CreateTemp("", "test.log")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(file.Name())
	}()

	tests := []struct {
		name    string
		config  configs.LogConfig
		eventFn func(*zerolog.Event) *zerolog.Event
		want    string
	}{
		{
			"success write log to file",
			configs.LogConfig{
				Level: "debug",
				Async: configs.AsyncLogConfig{
					Enabled: &falseVar,
				},
				Output: configs.OutputLogConfig{
					Type:     "file",
					Filename: file.Name(),
				}},
			func(e *zerolog.Event) *zerolog.Event {
				return e.Str("message", "test")
			},
			fmt.Sprintf("{\"message\":\"%s\",\"time\":\"%s\"}\n", "test", expectedTimestampString),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigureLogging(tt.config, NewProfiles())
			_ = file.Truncate(0) // clear file

			tt.eventFn(log.Log()).Send()

			b, err := os.ReadFile(file.Name())
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func Test_checkFileWriteable(t *testing.T) {
	readonlyDirectory := t.TempDir()
	err := os.Chmod(readonlyDirectory, os.FileMode(0444)) // perm: r--r--r--
	assert.NoError(t, err)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			"logging to temp directory should succeed",
			t.TempDir() + string(os.PathSeparator) + "valkyrie.log",
			true,
		},
		{
			"logging to read-only temp directory should fail",
			readonlyDirectory + string(os.PathSeparator) + "valkyrie.log",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, checkFileWriteable(tt.filename), "checkFileWriteable(%v)", tt.filename)
		})
	}
}
