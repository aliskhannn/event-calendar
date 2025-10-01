package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateLogger initializes and configures a new Zap logger instance.
// It sets up a production-ready JSON logger with an Info level, ISO8601 timestamp format,
// and includes the process ID in log entries. Logs are output to stdout, with errors to stderr.
//
// Returns:
//   - A pointer to the configured Zap logger.
func CreateLogger() *zap.Logger {
	// Configure encoder settings for JSON output.
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"                   // key for timestamp field in logs
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // format timestamps in ISO8601

	// Configure logger settings.
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel), // set minimum log level to Info
		Development:       false,                               // disable development mode
		DisableCaller:     false,                               // include caller information
		DisableStacktrace: false,                               // include stacktraces for errors
		Sampling:          nil,                                 // disable sampling
		Encoding:          "json",                              // use JSON encoding for logs
		EncoderConfig:     encoderCfg,                          // apply encoder configuration
		OutputPaths:       []string{"stdout"},                  // output logs to stdout
		ErrorOutputPaths:  []string{"stderr"},                  // output errors to stderr
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(), // include process ID in all log entries
		},
	}

	// Build and return the logger, panicking on failure.
	return zap.Must(config.Build())
}
