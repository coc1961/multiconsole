package command

import (
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
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	stdin   io.WriteCloser
	command *string
	cmd     *exec.Cmd
	err     error
	run     bool
}

//Stop stop
func (c *Command) Stop() error {
	c.run = false
	c.command = nil
	ret := c.cmd.Process.Kill()
	return ret
}

//Start start
func (c *Command) Start() (chan []byte, chan []byte) {
	c.run = true
	out0 := make(chan []byte, 1)
	out1 := make(chan []byte, 1)
	go func(c *Command) {
		cmd := exec.Command("sh", "-c", "/bin/sh", "--login")
		c.stdout, _ = cmd.StdoutPipe()
		c.stderr, _ = cmd.StderrPipe()
		c.stdin, _ = cmd.StdinPipe()
		c.cmd = cmd
		defer c.stdout.Close()
		defer c.stderr.Close()
		defer c.stdin.Close()

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func(c *Command, out chan<- []byte) {
			wgl := sync.WaitGroup{}
			for c.run {
				b := make([]byte, 10000)
				cont, err := c.stdout.Read(b)
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
						case <-time.After(10 * time.Millisecond):
							wgl.Done()
						}
					}(b1)
				}
			}
			wg.Done()
			wgl.Wait()
			close(out)
			//fmt.Print("SalgoHilo1 ")
		}(c, out0)

		go func(c *Command, out chan<- []byte) {
			wgl := sync.WaitGroup{}
			for c.run {
				b := make([]byte, 10000)
				cont, err := c.stderr.Read(b)
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
						case <-time.After(10 * time.Millisecond):
							wgl.Done()
						}
					}(b1)
				}
			}
			wg.Done()
			wgl.Wait()
			close(out)
			//fmt.Print("SalgoHilo2 ")
		}(c, out1)

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
		//close(out0)
		//close(out1)
		c.err = nil
		//fmt.Print("SALGO1 ")

	}(c)
	return out0, out1
}

//Execute Execute command
func (c *Command) Execute(cmd string) (int, error) {
	if !c.run {
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
