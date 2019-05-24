package main

import (
	"os"

	"github.com/coc1961/multiconsole/console"
)

func main() {
	//termbox.SetOutputMode(termbox.Output256)
	cv := console.NewConsoleView(os.Args[1:])
	cv.Start()
}
