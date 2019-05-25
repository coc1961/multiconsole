package console

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/coc1961/multiconsole/command"
	c "github.com/jroimartin/gocui"
	termbox "github.com/nsf/termbox-go"
)

//Consola consola
type Consola struct {
	command  *string
	cmd      *command.Command
	g        *c.Gui
	v        *c.View
	name     string
	x0       int
	y0       int
	x1       int
	y1       int
	running  bool
	commands map[string]Commands
	mutex    *sync.Mutex
}

//NewConsola NewConsola
func NewConsola(g *c.Gui, name string, x0, y0, x1, y1 int, cmd *string, commands map[string]Commands, mutex *sync.Mutex) *Consola {
	return &Consola{g: g, name: name, x0: x0, x1: x1, y0: y0, y1: y1, command: cmd, commands: commands, mutex: mutex}
}

//Execute Execute command
func (s *Consola) Execute(cmd string) (int, error) {
	if s.cmd.Run() {
		return s.cmd.Execute(cmd)
	}
	return -1, errors.New("Shell Closed")
}

//Error error
func (s *Consola) Error() error {
	return s.cmd.Error()
}

//Stop stop
func (s *Consola) Stop() error {
	ret := s.cmd.Stop()
	return ret
}

func (s *Consola) write(b []byte) (int, error) {
	if !s.cmd.Run() {
		return 0, nil
	}
	s.mutex.Lock()
	defer func() {
		s.mutex.Unlock()
	}()
	_, y := s.v.Size()
	if len(s.v.BufferLines()) > y*100 {
		s.v.Clear()
	}
	n, err := s.v.Write(b)
	termbox.Interrupt()
	return n, err
}

func (s *Consola) read() {
	s.v.Autoscroll = true
	s.running = true

	comm := s.cmd

	out0, out1 := comm.Start()
	for s.cmd.Run() {
		select {
		case b, ok := <-out0:
			if !ok {
				break
			}
			if _, err := s.write(b); err != nil {
				break
			}
		case b, ok := <-out1:
			if !ok {
				break
			}
			if _, err := s.write(b); err != nil {
				break
			}
		case <-time.After(100 * time.Millisecond):
		}
		//s.g.Update(s.updateConsole)
	}
	s.running = false
	s.v.Clear()
}

func (s *Consola) updateConsole(g *c.Gui) error {
	//termbox.Interrupt()
	/*
		s.mutex.Lock()
		defer s.mutex.Unlock()
	*/
	return nil
}

func (s *Consola) startCommand() {
	go s.read()
}

//Start start
func (s *Consola) Start() error {
	g := s.g
	s.cmd = command.NewCommand(s.command)
	s.command = nil

	/*
		if s.v != nil {
			s.v.Clear()
			s.startCommand()
			return nil
		}
	*/

	g.SetCurrentView(s.name + "View")
	if err := g.DeleteView(s.name + "View"); err != nil && err != c.ErrUnknownView {
		fmt.Println(err)
		return err
	}
	if err := g.DeleteView(s.name + "Input"); err != nil && err != c.ErrUnknownView {
		fmt.Println(err)
		return err
	}
	if v, err := g.SetView(s.name+"View", s.x0, s.y0, s.x1, s.y1-3); err != nil {
		if err != c.ErrUnknownView {
			fmt.Println(err)
			return err
		}
		s.v = v
		s.startCommand()
	}

	iv, err := g.SetView(s.name+"Input", s.x0, s.y1-2, s.x1, s.y1)
	if err != nil && err != c.ErrUnknownView {
		log.Println("Failed to create input view:", err)
		return err
	}
	iv.Title = "Input"
	iv.FgColor = c.ColorYellow
	// The input view shall be editable.
	iv.Editable = true
	err = iv.SetCursor(0, 0)
	if err != nil {
		log.Println("Failed to set cursor:", err)
		return err
	}

	// Make the enter key copy the input to the output.
	err = g.SetKeybinding(s.name+"Input", c.KeyEnter, c.ModNone, func(g *c.Gui, iv *c.View) error {
		iv.Rewind()

		ov, e := g.View(s.name + "View")
		if e != nil {
			log.Println("Cannot get output view:", e)
			return e
		}
		_, e = fmt.Fprint(ov, iv.Buffer())

		tmp := strings.Replace(iv.Buffer(), "\n", "", -1)
		c := s.commands[tmp]
		if c == nil {
			c = s.commands["default"]
		}
		c(s, tmp)

		if e != nil {
			log.Println("Cannot print to output view:", e)
		}
		iv.Clear()
		e = iv.SetCursor(0, 0)
		if e != nil {
			log.Println("Failed to set cursor:", e)
		}
		return e
	})

	err = g.SetKeybinding(s.name+"Input", c.KeyArrowDown, c.ModNone, s.historyDown)
	err = g.SetKeybinding(s.name+"Input", c.KeyArrowUp, c.ModNone, s.historyUp)

	_, err = g.SetCurrentView(s.name + "Input")
	if err != nil {
		log.Println("Cannot set focus to input view:", err)
		return err
	}

	return s.Error()
}

func (s *Consola) historyDown(g *c.Gui, iv *c.View) error {
	iv.Clear()
	iv.Rewind()
	iv.Write([]byte(s.cmd.HistoryNext()))
	return nil
}
func (s *Consola) historyUp(g *c.Gui, iv *c.View) error {
	iv.Clear()
	iv.Rewind()
	iv.Write([]byte(s.cmd.HistoryPrev()))
	return nil
}
