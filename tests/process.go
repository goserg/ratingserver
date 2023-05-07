package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

type Process struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewProcess(ctx context.Context, command string) *Process {
	return &Process{
		cmd:    exec.CommandContext(ctx, command),
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
}

func (p *Process) Start(ctx context.Context) error {
	startChan := make(chan error, 1)
	go func() {
		startChan <- p.cmd.Start()
	}()

	for {
		select {
		case err := <-startChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Process) Stop() (exitCode int, err error) {
	err = p.cmd.Process.Kill()
	if err != nil {
		return -1, errors.New("errors sending signal to process: " + err.Error())
	}
	state, err := p.cmd.Process.Wait()
	if err == nil {
		return -1, nil
	}
	return state.ExitCode(), err
}
