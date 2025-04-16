package logger

import (
	"fmt"
	"github.com/rifflock/lfshook"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化日志记录器
// mode 参数应传递 Gin 的当前模式（gin.DebugMode 或 gin.ReleaseMode）
func InitLogger(mode, logName string) (*logrus.Logger, error) {
	myLogger := logrus.New()

	// 设置日志格式为 TextFormatter，类似 Gin 的日志格式
	myLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05",
		ForceColors:      false,
		DisableColors:    false,
		DisableTimestamp: false,
		PadLevelText:     true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	// 根据 Gin 的模式设置日志级别
	if mode == gin.DebugMode {
		myLogger.SetLevel(logrus.DebugLevel)
	} else {
		myLogger.SetLevel(logrus.InfoLevel)
	}

	// 获取启动路径
	logBasePath := os.Getenv("LOG_BASE_PATH")
	if logBasePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %v", err)
		}
		logBasePath = cwd
	}

	// 确保日志目录存在
	logDir := filepath.Join(logBasePath, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %v", logDir, err)
	}

	// 控制台输出（仅 Debug 级别）
	consoleWriter := os.Stdout

	// 文件输出（Info 及以上级别），使用 lumberjack 实现基于大小的轮转
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logName),
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	}

	// 使用 lfshook 将不同级别的日志发送到不同的输出
	myLogger.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: consoleWriter,
			logrus.InfoLevel:  fileWriter,
			logrus.WarnLevel:  fileWriter,
			logrus.ErrorLevel: fileWriter,
			logrus.FatalLevel: fileWriter,
			logrus.PanicLevel: fileWriter,
		},
		myLogger.Formatter,
	))

	// 启用调用者信息
	myLogger.SetReportCaller(true)

	return myLogger, nil
}
