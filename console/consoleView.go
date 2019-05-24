package console

import (
	"fmt"
	"log"

	c "github.com/jroimartin/gocui"
)

//NewConsoleView NewConsoleView
func NewConsoleView(cmd []string) *ConsoleView {
	commands := getConsoleCommands()
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

	if err := g.SetKeybinding("", c.KeyTab, c.ModNone, cv.focus); err != nil {
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

func (cv *ConsoleView) focus(g *c.Gui, v *c.View) error {
	if _, err := g.SetCurrentView(fmt.Sprintf("cmd%dInput", cv.current)); err == nil {
		cv.current++
	}
	if cv.current > cv.maxConsole {
		cv.current = 1
	}
	return nil
}

func (cv *ConsoleView) quit(g *c.Gui, v *c.View) error {
	return c.ErrQuit
}

func (cv *ConsoleView) layout(g *c.Gui) error {
	if !cv.pri {
		return nil
	}
	cv.pri = false

	g.Cursor = true

	maxX, maxY := g.Size()

	_ = maxY
	if v, err := g.SetView("tit", 0, 0, maxX, 2); err != nil && err != c.ErrUnknownView {
		return err
	} else {
		v.Frame = false
		normal := "\033[0m"
		color := fmt.Sprintf("\033[3%d;%dm", 2, 2)
		color1 := fmt.Sprintf("\033[3%d;%dm", 2, 7)
		fmt.Fprintf(v, "%sMulti Consola%s - %sTab%s Cambio de Foco,  %sCtrl-c%s Salir,  %sEnter%s Ejecutar,  %sComandos %s exit,kill,cls", color1, normal, color, normal, color, normal, color, normal, color, normal)
	}

	newConsole(cv, 3, (maxX/2)+1, (maxY/2)+1, maxX-1, maxY-2)
	newConsole(cv, 2, 0, (maxY/2)+1, (maxX/2)-1, maxY-2)
	newConsole(cv, 1, (maxX/2)+1, 2, maxX-1, maxY/2)
	newConsole(cv, 0, 0, 2, (maxX/2)-1, maxY/2)

	return nil
}

func newConsole(cv *ConsoleView, ind, x, y, x1, y1 int) {
	{
		var cmd *string
		if len(cv.cmd) > ind {
			cmd = &cv.cmd[ind]
		}
		cmd1 := NewConsola(cv.g, fmt.Sprintf("cmd%d", ind+1), x, y, x1, y1, cmd, cv.commands)
		err := cmd1.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
}
