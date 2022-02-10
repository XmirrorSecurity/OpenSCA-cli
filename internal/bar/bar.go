/*
 * @Description: pargress bar
 * @Date: 2022-01-20 15:48:46
 */

package bar

import (
	"fmt"
	"opensca/internal/args"
)

var (
	id         int  = 0
	Dir        *Bar = newBar("scan dir")
	Archive    *Bar = newBar("unarchive")
	Maven      *Bar = newBar("parse maven indirect dependency")
	Npm        *Bar = newBar("parse npm indirect dependency")
	Composer   *Bar = newBar("parse composer indirect dependency")
	Dependency *Bar = newBar("parse project dependency")
)

// mult pargress bar
type Bar struct {
	text string
	now  int
	id   int
}

func newBar(text string) *Bar {
	return &Bar{
		id:   -1,
		now:  0,
		text: text,
	}
}

/**
 * @description: add progress
 * @param {int} n
 */
func (b *Bar) Add(n int) {
	if !args.ProgressBar {
		return
	}
	if b.id == -1 {
		id++
		b.id = id
		fmt.Println(b.text)
	}
	b.now += n
	fmt.Printf("\033[%dA\033[K", id-b.id+1)
	fmt.Printf("\r%s: %d", b.text, b.now)
	fmt.Printf("\033[%dB\033[K", id-b.id+1)
}
