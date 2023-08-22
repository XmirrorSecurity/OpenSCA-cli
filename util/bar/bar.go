package bar

import (
	"fmt"
	"sync"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/util/args"
)

var (
	id         int  = 0
	Dir        *Bar = newBar("scan dir")
	Archive    *Bar = newBar("unarchive")
	Maven      *Bar = newBar("parse maven indirect dependency")
	Npm        *Bar = newBar("parse npm indirect dependency")
	Composer   *Bar = newBar("parse composer indirect dependency")
	Dependency *Bar = newBar("parse project dependency")
	PipCompile *Bar = newBar("parse python module")
)

// mult pargress bar
type Bar struct {
	text string
	now  int
	id   int
	logo int
}

func newBar(text string) *Bar {
	return &Bar{
		id:   -1,
		now:  0,
		text: text,
	}
}

var (
	logos    = []string{`|`, `/`, `-`, `\`}
	logoOnce = sync.Once{}
	barlock  = sync.Mutex{}
	lastBar  *Bar
)

func updateLogo() {
	go func() {
		for {
			if lastBar != nil {
				lastBar.Add(0)
			}
			<-time.After(time.Millisecond * 100)
		}
	}()
}

// Add add progress
func (b *Bar) Add(n int) {
	if !args.Config.Bar {
		return
	}
	logoOnce.Do(updateLogo)
	barlock.Lock()
	defer barlock.Unlock()
	if b.id == -1 {
		id++
		b.id = id
		fmt.Println(b.text)
	}
	b.now += n
	fmt.Printf("\033[%dA\033[K", id-b.id+1)
	fmt.Printf("\r%s %s: %d", logos[b.logo], b.text, b.now)
	fmt.Printf("\033[%dB\033[K", id-b.id+1)
	b.logo++
	b.logo %= len(logos)
	lastBar = b
}
