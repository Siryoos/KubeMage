package main

import (
	"bufio"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type stdoutMsg struct {
	cmd string
	out string
}
type stderrMsg struct {
	cmd string
	out string
}
type execDoneMsg struct {
	cmd string
	err error
}

func execCmd(command string, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("bash", "-c", command)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error creating stdout pipe: %w", err)}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error creating stderr pipe: %w", err)}
		}

		if err := cmd.Start(); err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error starting command: %w", err)}
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				p.Send(stdoutMsg{cmd: command, out: scanner.Text()})
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				p.Send(stderrMsg{cmd: command, out: scanner.Text()})
			}
		}()

		err = cmd.Wait()
		return execDoneMsg{cmd: command, err: err}
	}
}
