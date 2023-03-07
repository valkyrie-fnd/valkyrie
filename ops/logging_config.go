package ops

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

// ConfigureLogging configures the logging framework
func ConfigureLogging(logConfig configs.LogConfig, profiles *Profiles) {
	level, err := zerolog.ParseLevel(logConfig.Level)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}
	// Profile to prevent changing of logging. Test will set up logging it self
	if profiles.Has("testlog") {
		return
	}
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return pkgerrors.MarshalStack(errors.WithStack(err))
	}
	// configure stack field name to stack_trace, that is automatically recognized by Google error reporting
	zerolog.ErrorStackFieldName = "stack_trace"

	// use custom json marshaller (goccy/go-json)
	zerolog.InterfaceMarshalFunc = json.Marshal

	// use RFC3339 with nano precision for timestamp field
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// configure output
	var writer io.Writer
	switch logConfig.Output.Type {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "file":
		writer = getFileWriter(logConfig.Output)
	default:
		writer = os.Stderr // default stderr to make sure it's not missed
	}
	log.Logger = log.Logger.Output(writer)

	// If async, wrap logger in a diode
	if logConfig.Async.Enabled != nil && *logConfig.Async.Enabled {
		log.Logger = log.Output(diode.NewWriter(writer, logConfig.Async.BufferSize, logConfig.Async.PollInterval, func(count int) {
			_, _ = fmt.Fprintf(os.Stderr, "Async logger buffer full, dropped %d messages\n", count)
		}))
	}

	// pretty logs for local use
	if profiles.Has("local") && logConfig.Output.Type != "file" {
		log.Logger = log.With().Caller().Logger().
			Output(zerolog.ConsoleWriter{Out: writer, TimeFormat: time.RFC3339Nano})
	}

	// use global logger as default context logger (used when context is missing a logger: "zerolog.Ctx(ctx).Info()")
	zerolog.DefaultContextLogger = &log.Logger

	log.Info().Strs("profiles", profiles.List()).Msg("Configured logging")
}

func getFileWriter(config configs.OutputLogConfig) io.Writer {
	if config.Filename == "" {
		config.Filename = "valkyrie.log"
	}

	// If async writing (diode) is enabled and the filename is not writeable, errors from lumberjack are not propagated
	// and results in logs being lost
	if !checkFileWriteable(config.Filename) {
		_, _ = fmt.Fprintf(os.Stderr, "Writing logs to stderr instead\n")
		return os.Stderr
	}

	return &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}
}

// checkFileWriteable checks if a filename (or path) is writable and returns true, otherwise logs an error
// to stderr and returns false.
//
// A filename is considered writable if:
// * directory either already exist, or is possible to create
// * filename can be created if missing
// * filename can be written to
func checkFileWriteable(filename string) bool {
	// get directory if any, "valkyrie.log" yields just "." as directory
	directory := path.Dir(filename)
	// perm=0744: owner: rwx, group: r, other: r
	// same permissions as lumberjack would create log directories with
	perm := os.FileMode(0744)
	err := os.MkdirAll(directory, perm)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create logging directory for '%s': %v\n", filename, err)
		return false
	}

	// perm=0644: owner: rw, group: r, other: r
	// same permissions as lumberjack would create log files with
	perm = os.FileMode(0644)
	// make sure we can create and write to the file
	flag := os.O_CREATE | os.O_WRONLY

	// #nosec G304
	file, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to open logging file for writing '%s': %v\n", filename, err)
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	return true
}
