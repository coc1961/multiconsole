package main

import (
	"os"

	"github.com/coc1961/multiconsole/console"
)

func main() {
	cv := console.NewConsoleView(os.Args[1:])
	cv.Start()
}
