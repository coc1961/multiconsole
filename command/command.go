package command

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

//NewCommand NewCommand
func NewCommand(cmd *string) *Command {
	return &Command{command: cmd}
}

//Command dos shell
type Command struct {
	stdin   io.WriteCloser
	cancel  context.CancelFunc
	command *string
	cmd     *exec.Cmd
	err     error
	run     bool
	running bool
}

//Stop stop
func (c *Command) Stop() error {
	c.run = false
	c.command = nil
	c.cancel()
	for c.IsRunning() {
		//fmt.Print(".")
		<-time.After(time.Millisecond * 200)
	}
	return nil
}

//Start start
func (c *Command) Start() (chan []byte, chan []byte) {
	c.run = true
	c.running = true
	out0 := make(chan []byte, 10000)
	out1 := make(chan []byte, 10000)
	go func(c *Command) {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		cmd := exec.CommandContext(ctx, "sh", "-c", "/bin/sh", "--login")
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		c.stdin, _ = cmd.StdinPipe()
		c.cmd = cmd

		wg := sync.WaitGroup{}
		wg.Add(2)

		go c.read(&wg, stdout, out0)
		go c.read(&wg, stderr, out1)

		c.err = cmd.Start()
		if c.err != nil {
			return
		}
		if c.command != nil {
			go func() {
				<-time.After(1 * time.Second)
				c.Execute(*c.command)
				c.Execute("\n")
				c.command = nil
			}()
		}

		cmd.Wait()
		//fmt.Print("SALGOCmd ")
		c.run = false

		wg.Wait()
		c.err = nil
		c.running = false

		//fmt.Print("SALGO1 ")

	}(c)
	return out0, out1
}

func (c *Command) read(wg *sync.WaitGroup, std io.ReadCloser, out chan []byte) {
	wgl := sync.WaitGroup{}
	for c.run {
		b := make([]byte, 10000)
		cont, err := std.Read(b)
		if err != nil {
			break
		}
		if cont > 0 {
			b1 := make([]byte, cont)
			copy(b1, b[0:cont])
			wgl.Add(1)
			go func(b []byte) {
				select {
				case out <- b:
					wgl.Done()
					return
				case <-time.After(10 * time.Millisecond):
					wgl.Done()
					return
				}
			}(b1)
		}
	}
	wg.Done()
	wgl.Wait()
	close(out)
	//fmt.Print("SalgoHilo1 ")
}

//Execute Execute command
func (c *Command) Execute(cmd string) (int, error) {
	if !c.run {
		return 0, errors.New("Shell Closed")
	}
	if c.stdin == nil {
		return 0, errors.New("Shell Closed")
	}
	return c.stdin.Write([]byte(cmd))
}

//Error error
func (c *Command) Error() error {
	return c.err
}

//IsRun error
func (c *Command) IsRun() bool {
	return c.run
}

//IsRunning error
func (c *Command) IsRunning() bool {
	return c.running
}
