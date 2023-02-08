package vplugin

import (
	"io"
	stdlog "log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// zerologAdapter acts as an adapter for hclog.Logger interface and zerolog.
//
// This is used to bridge logs sent over vplugin (using hclog) to zerolog that is
// used within Valkyrie.
type zerologAdapter struct {
	zlog *zerolog.Logger
}

func NewZerologAdapter() *zerologAdapter {
	return &zerologAdapter{zlog: &log.Logger}
}

func (z zerologAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	zlevel, err := zerolog.ParseLevel(level.String())
	if err != nil {
		zlevel = zerolog.DebugLevel
	}

	z.write(z.zlog.WithLevel(zlevel), msg, args)
}

func (z zerologAdapter) Trace(msg string, args ...interface{}) {
	z.write(z.zlog.Trace(), msg, args)
}

func (z zerologAdapter) Debug(msg string, args ...interface{}) {
	z.write(z.zlog.Debug(), msg, args)
}

func (z zerologAdapter) Info(msg string, args ...interface{}) {
	z.write(z.zlog.Info(), msg, args)
}

func (z zerologAdapter) Warn(msg string, args ...interface{}) {
	z.write(z.zlog.Warn(), msg, args)
}

func (z zerologAdapter) Error(msg string, args ...interface{}) {
	z.write(z.zlog.Error(), msg, args)
}

func (z zerologAdapter) IsTrace() bool {
	return z.zlog.GetLevel() <= zerolog.TraceLevel
}

func (z zerologAdapter) IsDebug() bool {
	return z.zlog.GetLevel() <= zerolog.DebugLevel
}

func (z zerologAdapter) IsInfo() bool {
	return z.zlog.GetLevel() <= zerolog.InfoLevel
}

func (z zerologAdapter) IsWarn() bool {
	return z.zlog.GetLevel() <= zerolog.WarnLevel
}

func (z zerologAdapter) IsError() bool {
	return z.zlog.GetLevel() <= zerolog.ErrorLevel
}

func (z zerologAdapter) ImpliedArgs() []interface{} {
	return []any{}
}

func (z zerologAdapter) With(_ ...interface{}) hclog.Logger {
	return z // noop
}

func (z zerologAdapter) Name() string {
	return "zerologAdapter"
}

func (z zerologAdapter) Named(_ string) hclog.Logger {
	return z // noop
}

func (z zerologAdapter) ResetNamed(_ string) hclog.Logger {
	return z // noop
}

func (z zerologAdapter) SetLevel(_ hclog.Level) {
	// noop
}

func (z zerologAdapter) GetLevel() hclog.Level {
	return hclog.LevelFromString(z.zlog.GetLevel().String())
}

func (z zerologAdapter) StandardLogger(_ *hclog.StandardLoggerOptions) *stdlog.Logger {
	panic("not implemented")
}

func (z zerologAdapter) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	panic("not implemented")
}

func (z zerologAdapter) write(event *zerolog.Event, msg string, args []any) {
	for i := 0; args != nil && i+1 < len(args); i += 2 {
		key := args[i].(string)
		switch key {
		case "timestamp":
			continue // skip timestamp field added by hclog, zerolog appends its own
		case "time":
			continue // skip timestamp field added by hclog, zerolog appends its own
		default:
			event.Interface(key, args[i+1])
		}
	}
	event.Msg(msg)
}
