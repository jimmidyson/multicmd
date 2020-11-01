package multicmd

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type MultiCmds struct {
	cmds []*exec.Cmd
}

func NewMultiCmds(cmds ...*exec.Cmd) *MultiCmds {
	return &MultiCmds{cmds: cmds}
}

func (c *MultiCmds) Start() error {
	var errs []error

	for _, cmd := range c.cmds {
		if err := cmd.Start(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", cmd, err))
		}
	}

	return newAggregate(errs)
}

func (c *MultiCmds) Wait() error {
	var (
		errs []error
		wg   sync.WaitGroup
		errc = make(chan error)
	)
	defer close(errc)

	for i := range c.cmds {
		cmd := c.cmds[i]

		wg.Add(1)
		go func() {
			if err := cmd.Wait(); err != nil {
				errc <- fmt.Errorf("%s: %v", cmd, err)
			} else {
				errc <- nil
			}
		}()
	}

	go func() {
		for err := range errc {
			if err != nil {
				errs = append(errs, err)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	return newAggregate(errs)
}

func (c *MultiCmds) SignalAll(sig os.Signal) error {
	var errs []error

	for i := range c.cmds {
		cmd := c.cmds[i]

		if cmd.Process != nil {
			if err := cmd.Process.Signal(sig); err != nil && err.Error() != "os: process already finished" {
				errs = append(errs, fmt.Errorf("%s: %v", cmd, err))
			}
		}
	}

	return newAggregate(errs)
}

func (c *MultiCmds) KillAll() error {
	var errs []error

	for i := range c.cmds {
		cmd := c.cmds[i]

		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil && err.Error() != "os: process already finished" {
				errs = append(errs, fmt.Errorf("%s: %v", cmd, err))
			}
		}
	}

	return newAggregate(errs)
}
