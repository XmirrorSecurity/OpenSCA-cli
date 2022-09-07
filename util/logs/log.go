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
	"strings"

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

func init() {
	// 创建日志文件
	var err error
	dir, _ := os.Executable()
	filepath := path.Join(path.Dir(strings.ReplaceAll(dir, `\`, `/`)), "opensca.log")
	logFile, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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
		return
	}
	logger.SetPrefix(fmt.Sprintf("[%s] ", prefixs[level]))
	logger.Output(3, fmt.Sprint(v))
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
