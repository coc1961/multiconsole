package console

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

func getConsoleCommands() map[string]Commands {
	commands := make(map[string]Commands)
	commands["exit"] = kill
	commands["kill"] = kill
	commands["cls"] = cls
	commands["clear"] = cls
	commands["default"] = execute
	return commands
}

func execute(s *Consola, cmd string) {
	s.Execute(cmd + "\n")
}

func cls(s *Consola, cmd string) {
	s.v.Clear()
	termbox.Interrupt()
}

func kill(s *Consola, cmd string) {
	s.Stop()
	go func() {
		for s.running {
			time.Sleep(100 * time.Millisecond)
		}
		s.Start()
		if s.Error() != nil {
			fmt.Println(s.Error())
		}
	}()
}
