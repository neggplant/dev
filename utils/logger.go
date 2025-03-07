package utils

import (
	"fmt"
	"os"
	"time"

	"myappg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

var logChan chan string

func InitAsyncLoggerFile() {
	logChan = make(chan string, 1000) // Buffered channel

	// Configure log output to file
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "logs/app.log", // Log file path
		MaxSize:    100,            // Maximum log file size (MB)
		MaxBackups: 3,              // Maximum number of old log files to keep
		MaxAge:     28,             // Maximum number of days to retain old log files
		Compress:   true,           // Whether to compress old log files
	}

	go func() {
		for logEntry := range logChan {
			_, err := lumberJackLogger.Write([]byte(logEntry + "\n"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing log entry: %v", err)
			}
		}
		lumberJackLogger.Close()
	}()

	// Custom time encoder
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05,000") + " - " + fmt.Sprintf("%d", os.Getpid()))
	}

	// Configure log encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(caller.File + " - " + fmt.Sprintf("%d", caller.Line))
	}
	encoderConfig.ConsoleSeparator = " - "

	// Implement a custom WriteSyncer
	ws := &chanWriteSyncer{
		LogChan: logChan,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		ws,
		config.GetLogLevel(),
	)

	// Add caller information
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	// Override the default zap.Logger
	zap.ReplaceGlobals(Logger)
}

type chanWriteSyncer struct {
	LogChan chan string
}

func (ws *chanWriteSyncer) Write(p []byte) (n int, err error) {
	ws.LogChan <- string(p)
	return len(p), nil
}

func (ws *chanWriteSyncer) Sync() error {
	return nil
}

func InitLoggerFile() {
	// Configure log output to file
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", // Log file path
		MaxSize:    100,            // Maximum log file size (MB)
		MaxBackups: 3,              // Maximum number of old log files to keep
		MaxAge:     28,             // Maximum number of days to retain old log files
		Compress:   true,           // Whether to compress old log files
	})

	// Custom time encoder
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05,000") + " - " + fmt.Sprintf("%d", os.Getpid()))
	}

	// Configure log encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(caller.File + " - " + fmt.Sprintf("%d", caller.Line))
	}
	encoderConfig.ConsoleSeparator = " - "

	// Create core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // Use ConsoleEncoder
		writer,                                   // Output to file
		config.GetLogLevel(),                     // Log level
	)

	// Add caller information
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	defer Logger.Sync()

	// Replace global logger and SugaredLogger
	zap.ReplaceGlobals(Logger)
}

func InitLogger() {
	var err error
	// Use production environment log configuration (JSON format, including call stack)
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Replace global logger and SugaredLogger
	zap.ReplaceGlobals(Logger)
}

func InitOriLoggerFile() {
	// Configure log output to file
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", // Log file path
		MaxSize:    100,            // Maximum log file size (MB)
		MaxBackups: 3,              // Maximum number of old log files to keep
		MaxAge:     28,             // Maximum number of days to retain old log files
		Compress:   true,           // Whether to compress old log files
	})

	// Configure log encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Time format

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON format
		writer,                                // Output to file
		config.GetLogLevel(),                  // Log level
	)

	// Create Logger
	Logger = zap.New(core, zap.AddCaller())
	defer Logger.Sync()

	// Replace global logger and SugaredLogger
	zap.ReplaceGlobals(Logger)
}
