package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dubbogo/gost/log/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig 日志配置
type LogConfig struct {
	Level    string // 日志级别: debug, info, warn, error, fatal
	Filename string // 日志文件路径
	MaxSize  int    // 单个日志文件最大大小(MB)
	MaxAge   int    // 日志文件保留天数
}

// GetLogConfigFromNacos 从 nacos 获取日志配置
func GetLogConfigFromNacos() (*LogConfig, error) {
	cfg := &LogConfig{
		Level:    "info", // 默认级别
		Filename: "",
		MaxSize:  100, // 默认 100MB
		MaxAge:   30,  // 默认保留 30 天
	}

	// 从 nacos 配置中读取日志配置
	if IsSet("log.level") {
		cfg.Level = GetString("log.level")
	}
	if IsSet("log.filename") {
		cfg.Filename = GetString("log.filename")
	}
	if IsSet("log.max_size") {
		cfg.MaxSize = GetInt("log.max_size")
	}
	if IsSet("log.max_age") {
		cfg.MaxAge = GetInt("log.max_age")
	}

	return cfg, nil
}

// InitLogger 初始化日志系统
func InitLogger(cfg *LogConfig) error {
	if cfg == nil {
		return fmt.Errorf("log config is nil")
	}

	// 1. 设置日志级别
	if !logger.SetLoggerLevel(cfg.Level) {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// 2. 如果配置了日志文件，设置文件输出
	if cfg.Filename != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// 使用 lumberjack 实现日志轮转
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize, // MB
			MaxBackups: 10,          // 最多保留10个备份文件
			MaxAge:     cfg.MaxAge,  // 保留天数
			Compress:   true,        // 压缩旧文件
			LocalTime:  true,        // 使用本地时间
		}

		// 解析日志级别
		var zapLevel zapcore.Level
		switch cfg.Level {
		case "debug":
			zapLevel = zapcore.DebugLevel
		case "info":
			zapLevel = zapcore.InfoLevel
		case "warn":
			zapLevel = zapcore.WarnLevel
		case "error":
			zapLevel = zapcore.ErrorLevel
		case "fatal":
			zapLevel = zapcore.FatalLevel
		default:
			zapLevel = zapcore.InfoLevel
		}

		// 创建 encoder 配置
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// 创建文件 core
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			zapLevel,
		)

		// 创建控制台 core (保留控制台输出)
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)

		// 合并两个 core
		core := zapcore.NewTee(fileCore, consoleCore)

		// 创建新的 logger
		zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

		// 设置全局 logger
		logger.SetLogger(zapLogger.Sugar())
	}

	logger.Infof("Logger initialized: level=%s, file=%s, max_size=%dMB, max_age=%d days",
		cfg.Level, cfg.Filename, cfg.MaxSize, cfg.MaxAge)

	return nil
}
