package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	c "github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

func main() {
	cv := NewConsoleView(os.Args[1:])
	cv.Start()
}

//NewConsoleView NewConsoleView
func NewConsoleView(cmd []string) *ConsoleView {
	cv := ConsoleView{pri: true, current: 2, maxConsole: 4, cmd: cmd}
	return &cv
}

//ConsoleView ConsoleView
type ConsoleView struct {
	cmd        []string
	pri        bool
	current    int
	g          *c.Gui
	maxConsole int
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
		fmt.Fprintf(v, "%sMulti Consola%s - %sTab%s Cambio de Foco,  %sCtrl-c%s Salir,  %sEnter%s Ejecutar,  %sTipear exit y presionar Enter%s Interrumpe proceso de consola", color1, normal, color, normal, color, normal, color, normal, color, normal)
	}

	{
		var cmd *string
		if len(cv.cmd) > 3 {
			cmd = &cv.cmd[3]
		}
		cmd1 := NewConsola(g, "cmd4", (maxX/2)+1, (maxY/2)+1, maxX-1, maxY-2, cmd)
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
		cmd1 := NewConsola(g, "cmd3", 0, (maxY/2)+1, (maxX/2)-1, maxY-2, cmd)
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
		cmd1 := NewConsola(g, "cmd2", (maxX/2)+1, 2, maxX-1, maxY/2, cmd)
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
		cmd1 := NewConsola(g, "cmd1", 0, 2, (maxX/2)-1, maxY/2, cmd)
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
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	stdin   io.WriteCloser
	command *string
	cmd     *exec.Cmd
	g       *c.Gui
	v       *c.View
	name    string
	err     error
	x0      int
	y0      int
	x1      int
	y1      int
	run     bool
}

//NewConsola NewConsola
func NewConsola(g *c.Gui, name string, x0, y0, x1, y1 int, cmd *string) *Consola {
	return &Consola{g: g, name: name, x0: x0, x1: x1, y0: y0, y1: y1, run: true, command: cmd}
}

//Execute Execute command
func (s *Consola) Execute(cmd string) (int, error) {
	return s.stdin.Write([]byte(cmd))
}

//Error error
func (s *Consola) Error() error {
	return s.err
}

//Stop stop
func (s *Consola) Stop() error {
	ret := s.cmd.Process.Kill()
	/*
		for s.run {
			<-time.After(100 * time.Millisecond)
		}
	*/
	return ret
}

func (s *Consola) write(b []byte) (int, error) {
	go func() {
		if r := recover(); r != nil {
			log.Println(r)
			return
		}
	}()
	if !s.run {
		return 0, nil
	}
	return s.v.Write(b)
}

//Start start
func (s *Consola) Start() error {
	g := s.g
	s.err = nil
	g.SetCurrentView(s.name + "View")
	if err := g.DeleteView(s.name + "View"); err != nil && err != c.ErrUnknownView {
		s.err = err
		return err
	}
	if v, err := g.SetView(s.name+"View", s.x0, s.y0, s.x1, s.y1-3); err != nil {
		if err != c.ErrUnknownView {
			s.err = err
			return err
		}
		s.v = v
		out := make(chan []byte, 10000)
		go func(s *Consola) {
			v.Autoscroll = true
			cmd := exec.Command("sh", "-c", "/bin/sh", "--login")
			s.stdout, _ = cmd.StdoutPipe()
			s.stderr, _ = cmd.StderrPipe()
			s.stdin, _ = cmd.StdinPipe()
			s.cmd = cmd
			s.run = true

			wg := sync.WaitGroup{}
			wg.Add(3)

			go func(s *Consola, out chan []byte) {
				for s.run {

					select {
					case b, ok := <-out:
						if !ok {
							break
						}
						_, err := s.write(b)
						if err != nil {
							break
						}
						termbox.Interrupt()
					case <-time.After(500 * time.Millisecond):
						if s.run {
							//fmt.Print(".")
						}
					}
				}
				wg.Done()
				//fmt.Print("Slgo1 ")
			}(s, out)

			go func(s *Consola, out chan []byte) {
				for s.run {
					b := make([]byte, 10000)
					cont, err := s.stdout.Read(b)
					if err != nil {
						break
					}
					if cont > 0 {
						out <- b[0:cont]
					}
				}
				wg.Done()
				//fmt.Print("Slgo2 ")
			}(s, out)
			go func(s *Consola, out chan []byte) {
				for s.run {
					b := make([]byte, 10000)
					cont, err := s.stderr.Read(b)
					if err != nil {
						break
					}
					if cont > 0 {
						out <- b[0:cont]
					}
				}
				wg.Done()
				//fmt.Print("Slgo3 ")
			}(s, out)
			s.err = cmd.Start()
			if s.err != nil {
				return
			}
			if s.command != nil {
				go func() {
					<-time.After(1 * time.Second)
					s.Execute(*s.command)
					s.Execute("\n")
					s.command = nil
				}()
			}

			cmd.Wait()
			s.run = false
			s.stderr.Close()
			s.stdout.Close()
			s.stdin.Close()
			//fmt.Println("Anter Wait")
			s.run = false
			wg.Wait()
			//fmt.Println("Despues Wait")
			close(out)
			s.run = true
			s.err = nil
		}(s)
	}

	iv, err := g.SetView(s.name+"Input", s.x0, s.y1-2, s.x1, s.y1)
	if err != nil && err != c.ErrUnknownView {
		log.Println("Failed to create input view:", err)
		s.err = err
		return err
	}
	iv.Title = "Input"
	iv.FgColor = c.ColorYellow
	// The input view shall be editable.
	iv.Editable = true
	err = iv.SetCursor(0, 0)
	if err != nil {
		log.Println("Failed to set cursor:", err)
		s.err = err
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

		if strings.Trim(iv.Buffer(), " ") != "exit\n" {
			s.Execute(iv.Buffer())
		} else {
			s.Stop()
			time.Sleep(300 * time.Millisecond)
			s.Start()
			if s.Error() != nil {
				fmt.Println(s.Error())
			}

		}

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
		s.err = err
		return err
	}

	return s.err
}
