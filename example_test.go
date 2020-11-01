package multicmd_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/jimmidyson/multicmd"
)

func ExampleWithTimeouts() {
	cmd := exec.Command("sleep", "5s")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := multicmd.StartAndWaitOrStop(ctx, cmds, os.Interrupt, 0); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cmd.ProcessState.ExitCode())

	// Output:
	// context deadline exceeded
	// -1
}

func ExampleWithTimeoutsAndMultipleCommands() {
	cmd1 := exec.Command("sleep", "5s")
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	cmd2 := exec.Command("sleep", "1s")
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd1, cmd2)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := multicmd.StartAndWaitOrStop(ctx, cmds, os.Interrupt, 0); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cmd1.ProcessState.ExitCode())
	fmt.Println(cmd2.ProcessState.ExitCode())

	// Output:
	// context deadline exceeded
	// -1
	// 0
}

func ExampleWithTimeoutsAndMultipleCommandsKillOnly() {
	cmd1 := exec.Command("sleep", "5s")
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	cmd2 := exec.Command("sleep", "1s")
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd1, cmd2)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := multicmd.StartAndWaitOrStop(ctx, cmds, nil, 1); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cmd1.ProcessState.ExitCode())
	fmt.Println(cmd2.ProcessState.ExitCode())

	// Output:
	// /usr/bin/sleep 5s: signal: killed
	// -1
	// 0
}

func ExampleWithTimeoutsAndMultipleCommandsWithInterruptAndKill() {
	cmd1 := exec.Command("sleep", "5s")
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	cmd2 := exec.Command("sleep", "1s")
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd1, cmd2)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := multicmd.StartAndWaitOrStop(ctx, cmds, os.Interrupt, 1); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cmd1.ProcessState.ExitCode())
	fmt.Println(cmd2.ProcessState.ExitCode())

	// Output:
	// context deadline exceeded
	// -1
	// 0
}

func ExampleWithSignalAll() {
	cmd1 := exec.Command("sleep", "5s")
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	cmd2 := exec.Command("sleep", "1s")
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd1, cmd2)

	if err := cmds.Start(); err != nil {
		_ = cmds.KillAll()
		fmt.Println(err)
		return
	}

	<-time.After(2 * time.Second)

	if err := cmds.SignalAll(os.Interrupt); err != nil {
		_ = cmds.KillAll()
		fmt.Println(err)
		return
	}

	if err := cmds.Wait(); err != nil {
		_ = cmds.KillAll()
		fmt.Println(err)
	}

	fmt.Println(cmd1.ProcessState.ExitCode())
	fmt.Println(cmd2.ProcessState.ExitCode())

	// Output:
	// /usr/bin/sleep 5s: signal: interrupt
	// -1
	// 0
}

func ExamplePropagateSignals() {
	cmd1 := exec.Command("sleep", "5s")
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	cmd2 := exec.Command("sleep", "1s")
	cmd2.Stderr = os.Stderr
	cmd2.Stdout = os.Stdout

	cmds := multicmd.NewMultiCmds(cmd1, cmd2)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := multicmd.StartAndWaitWithPropagatedSignals(cmds); err != nil {
			fmt.Println(err)
		}
		fmt.Println(cmd1.ProcessState.ExitCode())
		fmt.Println(cmd2.ProcessState.ExitCode())
	}()

	<-time.After(2 * time.Second)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
	wg.Wait()

	// Output:
	// /usr/bin/sleep 5s: signal: interrupt
	// -1
	// 0
}
