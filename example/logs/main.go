package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

func main() {

	// DEBUG log
	logs.Debugf("this is debug message %s", "hello world")

	// set log config
	logs.SetLogConfig(func(n *logs.LogConfig) {
		// close DEBUG log
		n.Debug = false
	})

	// this message will be ignore
	logs.Debug("this is debug message")

	logs.Infof("this is info message %d", 123)

	// WARN log
	logs.Warn("warn!")

	// ERROR log will print error stack
	logs.Error(errors.New("test error"))

	// custom log format
	logs.RegisterOut(func(level logs.Level, format string, v ...any) {
		log.SetPrefix(fmt.Sprintf("{CUSTOM} level:%d ", level))
		if format == "" {
			log.Output(3, fmt.Sprint(v...))
		} else {
			log.Output(3, fmt.Sprintf(format, v...))
		}
	})

	logs.Debug("custom log")

}
