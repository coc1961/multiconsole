package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	c "github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

func main() {
	cv := NewConsoleView()
	cv.Start()
}

//NewConsoleView NewConsoleView
func NewConsoleView() *ConsoleView {
	cv := ConsoleView{pri: true, current: 2, maxConsole: 4}
	return &cv
}

//ConsoleView ConsoleView
type ConsoleView struct {
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
		cmd1 := NewConsola(g, "cmd4", (maxX/2)+1, (maxY/2)+1, maxX-1, maxY-2)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		cmd1 := NewConsola(g, "cmd3", 0, (maxY/2)+1, (maxX/2)-1, maxY-2)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		cmd1 := NewConsola(g, "cmd2", (maxX/2)+1, 2, maxX-1, maxY/2)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	{
		cmd1 := NewConsola(g, "cmd1", 0, 2, (maxX/2)-1, maxY/2)
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
	stdout io.ReadCloser
	stderr io.ReadCloser
	stdin  io.WriteCloser
	cmd    *exec.Cmd
	g      *c.Gui
	v      *c.View
	name   string
	x0     int
	y0     int
	x1     int
	y1     int
}

//NewConsola NewConsola
func NewConsola(g *c.Gui, name string, x0, y0, x1, y1 int) *Consola {
	return &Consola{g: g, name: name, x0: x0, x1: x1, y0: y0, y1: y1}
}

//Stop stop
func (s *Consola) Stop() error {
	return s.cmd.Process.Kill()
}

//Start start
func (s *Consola) Start() error {
	g := s.g
	if err := g.DeleteView(s.name + "View"); err != nil && err != c.ErrUnknownView {
		return err
	}
	if v, err := g.SetView(s.name+"View", s.x0, s.y0, s.x1, s.y1-3); err != nil {
		if err != c.ErrUnknownView {
			return err
		}
		s.v = v
		out := make(chan []byte)
		go func(s *Consola) {
			v.Autoscroll = true
			cmd := exec.Command("sh", "-c", "/bin/sh")
			s.stdout, _ = cmd.StdoutPipe()
			s.stderr, _ = cmd.StderrPipe()
			s.stdin, _ = cmd.StdinPipe()
			s.cmd = cmd
			cmd.Start()

			go func(s *Consola, out chan []byte) {
				for {
					b, ok := <-out
					if !ok {
						return
					}
					s.v.Write(b)
					termbox.Interrupt()
				}
			}(s, out)

			go func(s *Consola, out chan []byte) {
				for true {
					b := make([]byte, 100)
					cont, err := s.stdout.Read(b)
					if err != nil {
						break
					}
					if cont > 0 && b != nil && len(b) >= cont {
						out <- b[0:cont]
					}
				}
			}(s, out)
			go func(s *Consola, out chan []byte) {
				for true {
					b := make([]byte, 100)
					cont, err := s.stderr.Read(b)
					if err != nil {
						break
					}
					if cont > 0 && b != nil && len(b) >= cont {
						out <- b[0:cont]
					}
				}
			}(s, out)
			cmd.Wait()
			close(out)
		}(s)
	}

	iv, err := g.SetView(s.name+"Input", s.x0, s.y1-2, s.x1, s.y1)
	if err != nil && err != c.ErrUnknownView {
		log.Println("Failed to create input view:", err)
		return nil
	}
	iv.Title = "Input"
	iv.FgColor = c.ColorYellow
	// The input view shall be editable.
	iv.Editable = true
	err = iv.SetCursor(0, 0)
	if err != nil {
		log.Println("Failed to set cursor:", err)
		return nil
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
			s.stdin.Write([]byte(iv.Buffer()))
		} else {
			s.Stop()
			s.Start()
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
	}

	return nil
}
