package logs

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
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
	execfile, _ := os.Executable()
	logDir := path.Join(path.Dir(execfile), "log")
	logPre := "log-"
	if f, err := os.Stat(logDir); err != nil {
		if err = os.Mkdir(logDir, 0755); err != nil {
			Warnf("create %s error: %v", logDir, err)
			return
		}
	} else {
		if !f.IsDir() {
			Warnf("%s is not a directory", logDir)
			return
		}
	}
	fs, err := os.ReadDir(logDir)
	if err != nil {
		Warnf("read %s err: %v", logDir, err)
		return
	}
	nums := []int{} // 记录日志编号
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), logPre) {
			if n, err := strconv.Atoi(f.Name()[len(logPre):]); err == nil {
				nums = append(nums, n)
			}
		}
	}
	sort.Ints(nums)
	id := 1
	if len(nums) > 0 {
		id = nums[len(nums)-1] + 1
	}
	f, err := os.OpenFile(path.Join(logDir, fmt.Sprintf("log-%d", id)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Warnf("open logfile err: %v", err)
	} else {
		log.SetOutput(f)
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
