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

	// 配置日志输出到文件
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "logs/app.log", // 日志文件路径
		MaxSize:    100,            // 日志文件最大大小（MB）
		MaxBackups: 3,              // 保留的旧日志文件最大数量
		MaxAge:     28,             // 保留旧日志文件的最大天数
		Compress:   true,           // 是否压缩旧日志文件
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

	// 配置日志编码器
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

	// 创建核心
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
	// 配置日志输出到文件
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", // 日志文件路径
		MaxSize:    100,            // 日志文件最大大小（MB）
		MaxBackups: 3,              // 保留的旧日志文件最大数量
		MaxAge:     28,             // 保留旧日志文件的最大天数
		Compress:   true,           // 是否压缩旧日志文件
	})

	// Custom time encoder
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05,000") + " - " + fmt.Sprintf("%d", os.Getpid()))
	}

	// 配置日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(caller.File + " - " + fmt.Sprintf("%d", caller.Line))
	}
	encoderConfig.ConsoleSeparator = " - "

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 使用ConsoleEncoder
		writer,                                   // 输出到文件
		config.GetLogLevel(),                     // 日志级别
	)

	// Add caller information
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	defer Logger.Sync()

	// 替换全局的日志器和 SugaredLogger
	zap.ReplaceGlobals(Logger)
}

func InitLogger() {
	var err error
	// 使用生产环境的日志配置（JSON 格式，包含调用堆栈）
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 替换全局的日志器和 SugaredLogger
	zap.ReplaceGlobals(Logger)
}

func InitOriLoggerFile() {
	// 配置日志输出到文件
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", // 日志文件路径
		MaxSize:    100,            // 日志文件最大大小（MB）
		MaxBackups: 3,              // 保留的旧日志文件最大数量
		MaxAge:     28,             // 保留旧日志文件的最大天数
		Compress:   true,           // 是否压缩旧日志文件
	})

	// 配置日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 时间格式

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON 格式
		writer,                                // 输出到文件
		config.GetLogLevel(),                  // 日志级别
	)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller())
	defer Logger.Sync()

	// 替换全局的日志器和 SugaredLogger
	zap.ReplaceGlobals(Logger)
}
