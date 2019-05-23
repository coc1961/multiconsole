package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coc1961/multiconsole/command"
	c "github.com/jroimartin/gocui"
)

func main() {
	//termbox.SetOutputMode(termbox.Output256)
	cv := NewConsoleView(os.Args[1:])
	cv.Start()
}

//NewConsoleView NewConsoleView
func NewConsoleView(cmd []string) *ConsoleView {
	commands := make(map[string]Commands)
	commands["exit"] = kill
	commands["kill"] = kill
	commands["cls"] = cls
	commands["clear"] = cls
	commands["default"] = execute

	cv := ConsoleView{pri: true, current: 2, maxConsole: 4, cmd: cmd, commands: commands}
	return &cv
}

//Commands Commands
type Commands func(s *Consola, cmd string)

//ConsoleView ConsoleView
type ConsoleView struct {
	cmd        []string
	pri        bool
	current    int
	g          *c.Gui
	maxConsole int
	commands   map[string]Commands
}

//Start start
func (cv *ConsoleView) Start() error {
	g, err := c.NewGui(c.OutputNormal)
	if err != nil {
		log.Println("Failed to create a GUI:", err)
		return nil
	}
	defer g.Close()

	g.SetManagerFunc(cv.layout)
	cv.g = g

	if err := g.SetKeybinding("", c.KeyTab, c.ModNone, func(g *c.Gui, v *c.View) error {
		if _, err := g.SetCurrentView(fmt.Sprintf("cmd%dInput", cv.current)); err == nil {
			cv.current++
		}
		if cv.current > cv.maxConsole {
			cv.current = 1
		}
		return nil
	}); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", c.KeyCtrlC, c.ModNone, cv.quit); err != nil {
		log.Panicln(err)
	}

	if err := cv.g.MainLoop(); err != nil && err != c.ErrQuit {
		log.Panicln(err)
		return err
	}
	return nil
}

func (cv *ConsoleView) layout(g *c.Gui) error {
	if !cv.pri {
		return nil
	}
	cv.pri = false

	g.Cursor = true

	maxX, maxY := g.Size()

	_ = maxY
	if v, err := g.SetView("tit", 0, 0, maxX, 2); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
		v.Frame = false

		normal := "\033[0m"
		color := fmt.Sprintf("\033[3%d;%dm", 2, 2)
		color1 := fmt.Sprintf("\033[3%d;%dm", 2, 7)
		fmt.Fprintf(v, "%sMulti Consola%s - %sTab%s Cambio de Foco,  %sCtrl-c%s Salir,  %sEnter%s Ejecutar,  %sComandos %s exit,kill,cls", color1, normal, color, normal, color, normal, color, normal, color, normal)
	}

	{
		var cmd *string
		if len(cv.cmd) > 3 {
			cmd = &cv.cmd[3]
		}
		cmd1 := NewConsola(g, "cmd4", (maxX/2)+1, (maxY/2)+1, maxX-1, maxY-2, cmd, cv.commands)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		var cmd *string
		if len(cv.cmd) > 2 {
			cmd = &cv.cmd[2]
		}
		cmd1 := NewConsola(g, "cmd3", 0, (maxY/2)+1, (maxX/2)-1, maxY-2, cmd, cv.commands)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		var cmd *string
		if len(cv.cmd) > 1 {
			cmd = &cv.cmd[1]
		}
		cmd1 := NewConsola(g, "cmd2", (maxX/2)+1, 2, maxX-1, maxY/2, cmd, cv.commands)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		var cmd *string
		if len(cv.cmd) > 0 {
			cmd = &cv.cmd[0]
		}
		cmd1 := NewConsola(g, "cmd1", 0, 2, (maxX/2)-1, maxY/2, cmd, cv.commands)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (cv *ConsoleView) quit(g *c.Gui, v *c.View) error {
	return c.ErrQuit
}

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
}

//NewConsola NewConsola
func NewConsola(g *c.Gui, name string, x0, y0, x1, y1 int, cmd *string, commands map[string]Commands) *Consola {
	return &Consola{g: g, name: name, x0: x0, x1: x1, y0: y0, y1: y1, command: cmd, commands: commands}
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
	return s.v.Write(b)
}

func (s *Consola) read() {
	s.v.Autoscroll = true
	//s.v.FgColor = c.ColorYellow + c.AttrBold
	s.running = true

	comm := s.cmd

	out0, out1 := comm.Start()
	for s.cmd.Run() {
		select {
		case b, ok := <-out0:
			if !ok {
				break
			}

			//fmt.Print(string(b))

			_, err := s.write(b)
			if err != nil {
				break
			}
			/*
				if l != len(b) {
					fmt.Println("MALLLLLLLLLLLLLLLLLLL")
				}
			*/
			s.g.Update(updateConsole)
			//termbox.Interrupt()
		case b, ok := <-out1:
			if !ok {
				break
			}
			_, err := s.write(b)
			if err != nil {
				break
			}
			s.g.Update(updateConsole)
			//termbox.Interrupt()
		case <-time.After(100 * time.Millisecond):
			s.g.Update(updateConsole)
			//termbox.Interrupt()
			//fmt.Print(s.cmd.IsRun(), " ")
			//termbox.SetCell(0, 0, '█', termbox.ColorBlack, termbox.ColorCyan)

		}
	}
	s.running = false
	s.v.Clear()
	//fmt.Print("SALGO0 ")
}

func updateConsole(g *c.Gui) error {
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

	if s.v != nil {
		s.startCommand()
		return nil
	}

	g.SetCurrentView(s.name + "View")
	if err := g.DeleteView(s.name + "View"); err != nil && err != c.ErrUnknownView {
		return err
	}
	if err := g.DeleteView(s.name + "Input"); err != nil && err != c.ErrUnknownView {
		return err
	}
	if v, err := g.SetView(s.name+"View", s.x0, s.y0, s.x1, s.y1-3); err != nil {
		if err != c.ErrUnknownView {
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

	_, err = g.SetCurrentView(s.name + "Input")
	if err != nil {
		log.Println("Cannot set focus to input view:", err)
		return err
	}

	return s.Error()
}

func execute(s *Consola, cmd string) {
	s.Execute(cmd + "\n")
}
func cls(s *Consola, cmd string) {
	s.v.Clear()
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
