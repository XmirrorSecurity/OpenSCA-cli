/*
 * @Descripation:
 * @Date: 2021-11-06 15:44:20
 */

package logs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"util/args"

	"github.com/pkg/errors"
)

type logLevel int

const (
	levelDebug logLevel = iota
	levelInfo
	levelWarning
	levelError
)

var (
	prefixs = []string{
		"DEBUG",
		"INFO",
		"WARNING",
		"ERROR",
	}
	logFile *os.File
	logger  *log.Logger
)

func InitLogger() {
	// 创建日志文件
	var err error

	cwd, err := os.Getwd() //获取当前工作目录
	if err != nil {
		fmt.Println(err)
	}

	logFilePath := args.Config.Logfile
	// fmt.Printf("Log file initLogger:%s\n", args.Config.Logfile)

	if logFilePath == "" {
		logFilePath = path.Join(strings.ReplaceAll(cwd, `\`, `/`), "opensca.log")
	}

	fmt.Printf("log file: %s\n", logFilePath)

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		fmt.Println(err)
		return
	}

	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Println("log file create fail!")
	} else {
		// 创建日志
		logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

func GetLogFile() *os.File {
	return logFile
}

func out(level logLevel, v interface{}) {
	if logger == nil {
		InitLogger()
	}
	logger.SetPrefix(fmt.Sprintf("[%s] ", prefixs[level]))
	err := logger.Output(3, fmt.Sprint(v))
	if err != nil {
		fmt.Println(errors.WithStack(err).Error())
		return
	}
}

func Debug(v interface{}) {
	out(levelDebug, v)
}

func Info(v interface{}) {
	out(levelInfo, v)
}

func Warn(v interface{}) {
	out(levelWarning, v)
}

func Error(err error) {
	out(levelError, errors.WithStack(err).Error())
}
