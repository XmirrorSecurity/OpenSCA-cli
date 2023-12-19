package main

import (
	"fmt"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func main() {
	file := model.NewFile("./test_file.txt", "")
	lineFunc := func(line string) { fmt.Println(line) }
	fmt.Println("read file by line:")
	file.ReadLine(lineFunc)
	fmt.Println("\nread file by line, no comment c type:")
	file.ReadLineNoComment(model.CTypeComment, lineFunc)
	fmt.Println("\nread file by line, no comment python type:")
	file.ReadLineNoComment(model.PythonTypeComment, lineFunc)
}
