package multicmd

import (
	"context"
	"os"
	"time"
)

func StartAndWaitOrStop(ctx context.Context, cmds *MultiCmds, interrupt os.Signal, killDelay time.Duration) error {
	if err := cmds.Start(); err != nil {
		_ = cmds.KillAll()
		return err
	}

	errc := make(chan error)
	go func() {
		select {
		case errc <- nil:
			return
		case <-ctx.Done():
		}

		var err error
		if interrupt != nil {
			if err = cmds.SignalAll(interrupt); err == nil {
				err = ctx.Err()
			}
		}

		if killDelay > 0 {
			timer := time.NewTimer(killDelay)
			select {
			// Report ctx.Err() as the reason we interrupted the processes...
			case errc <- ctx.Err():
				if !timer.Stop() {
					<-timer.C
				}
				return
			// ...but after killDelay has elapsed, fall back to a stronger signal.
			case <-timer.C:
			}

			// Wait still hasn't returned.
			// Kill the processes harder to make sure that it exits.
			//
			// Ignore any error: if cmd.Process has already terminated, we still
			// want to send ctx.Err() (or the error from the Interrupt call)
			// to properly attribute the signal that may have terminated it.
			_ = cmds.KillAll()
		}

		errc <- err
	}()

	waitErr := cmds.Wait()

	if interruptErr := <-errc; interruptErr != nil {
		return interruptErr
	}

	return waitErr
}
