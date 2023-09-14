package logs

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime/debug"
)

type LogConfig struct {
	Trace bool `json:"trace"`
	Debug bool `json:"debug"`
	Info  bool `json:"info"`
	Warn  bool `json:"warn"`
	Error bool `json:"error"`
}

type Level int8

const (
	_TRACE Level = iota
	_DEBUG
	_INFO
	_WARN
	_ERROR
)

var (
	prefix = map[Level]string{
		_TRACE: "[TRACE] ",
		_DEBUG: "[DEBUG] ",
		_INFO:  "[INFO] ",
		_WARN:  "[WARN] ",
		_ERROR: "[ERROR] ",
	}
	config = DefalutLogConfig()
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func DefalutLogConfig() LogConfig {
	return LogConfig{
		Trace: false,
		Debug: true,
		Info:  true,
		Warn:  true,
		Error: true,
	}
}

func CreateLog(logPath string) {

	if logPath == "" {
		Debug("create default log")
		createDefaultLog()
		return
	}

	if f, err := os.Create(logPath); err != nil {
		Warnf("create log %s err: %s, create default log", logPath, err)
		createDefaultLog()
	} else {
		Debugf("log file: %s", logPath)
		log.SetOutput(f)
	}

}

func createDefaultLog() {

	const logname = "opensca.log"
	defaultLogPaths := []string{}

	// 在工作目录生成日志
	if p, err := os.Getwd(); err == nil {
		defaultLogPaths = append(defaultLogPaths, filepath.Join(p, logname))
	}

	// 在cli目录生成日志
	if execfile, err := os.Executable(); err == nil {
		defaultLogPaths = append(defaultLogPaths, filepath.Join(filepath.Dir(execfile), logname))
	}

	// 在用户目录生成日志
	if user, err := user.Current(); err == nil {
		defaultLogPaths = append(defaultLogPaths, filepath.Join(user.HomeDir, logname))
	}

	for _, p := range defaultLogPaths {
		f, err := os.Create(p)
		if err != nil {
			Warnf("create log %s err: %s", p, err)
		} else {
			Debugf("log file: %s", p)
			log.SetOutput(f)
			break
		}
	}

}

// RegisterOut 注册日志输出格式
func RegisterOut(outFunc func(level Level, format string, v ...any)) {
	if outFunc != nil {
		out = outFunc
	}
}

var out = func(level Level, format string, v ...any) {
	if (level == _TRACE && !config.Trace) ||
		(level == _DEBUG && !config.Debug) ||
		(level == _INFO && !config.Info) ||
		(level == _WARN && !config.Warn) ||
		(level == _ERROR && !config.Error) {
		return
	}
	log.SetPrefix(prefix[level])
	if format == "" {
		log.Output(3, fmt.Sprint(v...))
	} else {
		log.Output(3, fmt.Sprintf(format, v...))
	}
}

func Trace(v ...any) {
	out(_TRACE, "", v...)
}

func Tracef(format string, v ...any) {
	out(_TRACE, format, v...)
}

func Debug(v ...any) {
	out(_DEBUG, "", v...)
}

func Debugf(format string, v ...any) {
	out(_DEBUG, format, v...)
}

func Info(v ...any) {
	out(_INFO, "", v...)
}

func Infof(format string, v ...any) {
	out(_INFO, format, v...)
}

func Warn(v ...any) {
	out(_WARN, "", v...)
}

func Warnf(format string, v ...any) {
	out(_WARN, format, v...)
}

func Error(v ...any) {
	out(_ERROR, "", v...)
	out(_ERROR, "", string(debug.Stack()))
}

func Errorf(format string, v ...any) {
	out(_ERROR, format, v...)
	out(_ERROR, "", string(debug.Stack()))
}

func Recover() {
	if err := recover(); err != nil {
		log.SetPrefix(prefix[_ERROR])
		log.Output(5, fmt.Sprint(err))
		log.Output(5, string(debug.Stack()))
	}
}
