package main

import (
	"bufio"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type execMsg string
type execDoneMsg struct{}

func execCmd(command string, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("bash", "-c", command)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		cmd.Start()

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				p.Send(execMsg(scanner.Text()))
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				p.Send(execMsg(scanner.Text()))
			}
		}()

		cmd.Wait()
		return execDoneMsg{}
	}
}
