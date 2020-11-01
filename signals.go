package multicmd

import (
	"fmt"
	"os"
	"os/signal"
)

func StartAndWaitWithPropagatedSignals(cmds *MultiCmds, interrupts ...os.Signal) error {
	c := make(chan os.Signal, 1)
	defer close(c)
	signal.Notify(c, interrupts...)
	defer signal.Stop(c)

	go func() {
		for sig := range c {
			if err := cmds.SignalAll(sig); err != nil {
				fmt.Fprintf(os.Stderr, "failed to relay signal %s to commands: %v\n", sig.String(), err)
			}
		}
	}()

	if err := cmds.Start(); err != nil {
		_ = cmds.KillAll()
		return err
	}

	return cmds.Wait()
}
