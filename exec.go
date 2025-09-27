package main

import (
	"bufio"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type execMsg string
type execDoneMsg struct{}

func execCmd(command string, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("bash", "-c", command)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			p.Send(execMsg(fmt.Sprintf("error creating stdout pipe: %v", err)))
			return execDoneMsg{}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			p.Send(execMsg(fmt.Sprintf("error creating stderr pipe: %v", err)))
			return execDoneMsg{}
		}

		if err := cmd.Start(); err != nil {
			p.Send(execMsg(fmt.Sprintf("error starting command: %v", err)))
			return execDoneMsg{}
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				p.Send(execMsg(scanner.Text()))
			}
			if err := scanner.Err(); err != nil {
				p.Send(execMsg(fmt.Sprintf("stdout error: %v", err)))
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				p.Send(execMsg(scanner.Text()))
			}
			if err := scanner.Err(); err != nil {
				p.Send(execMsg(fmt.Sprintf("stderr error: %v", err)))
			}
		}()

		if err := cmd.Wait(); err != nil {
			p.Send(execMsg(fmt.Sprintf("command exited with error: %v", err)))
		}
		return execDoneMsg{}
	}
}
