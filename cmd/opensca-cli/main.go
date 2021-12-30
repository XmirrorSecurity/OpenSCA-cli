/*
 * @Descripation: 引擎入口
 * @Date: 2021-11-03 11:12:19
 */
package main

import (
	"flag"
	"opensca/internal/args"
	"opensca/internal/engine"
)

func main() {
	args.Parse()
	if len(args.Filepath) > 0 {
		engine.NewEngine().ParseFile(args.Filepath)
	} else {
		flag.PrintDefaults()
	}
}
