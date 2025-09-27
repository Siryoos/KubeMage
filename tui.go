package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	userStyle     lipgloss.Style
	assistStyle   lipgloss.Style
	execStyle     lipgloss.Style
	errorStyle    lipgloss.Style
	viewportStyle lipgloss.Style
	headerStyle   lipgloss.Style
}

func defaultStyles() styles {
	return styles{
		userStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),  // Blue
		assistStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")),   // Green
		execStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")), // Yellow
		errorStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")), // Red
		viewportStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1),
	}
}

const (
	user   = "User"
	assist = "Assistant"
	execSender   = "Exec"
)

type ollamaStreamMsg string
type ollamaStreamDoneMsg struct{}

func generateStreamCmd(prompt string, p *tea.Program, model string) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan string)
		go GenerateStream(prompt, ch, model)

		// This will block until the first message is received
		initialResponse, ok := <-ch
		if !ok {
			return ollamaStreamDoneMsg{}
		}

		// Start a new goroutine to continue receiving messages
		go func() {
			for response := range ch {
				p.Send(ollamaStreamMsg(response))
			}
			p.Send(ollamaStreamDoneMsg{})
		}()

		return ollamaStreamMsg(initialResponse)
	}
}

type message struct {
	sender  string
	content string
}

type model struct {
	program     *tea.Program
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []message
	sender      string
	command     string
	ollamaModel string
	styles      styles
}

func InitialModel() *model {
	ta := textarea.New()
	ta.Placeholder = "Ask KubeMage a question..."
	ta.Focus()

	vp := viewport.New(80, 20)
	styles := defaultStyles()
	vp.Style = styles.viewportStyle

	return &model{
		textarea:    ta,
		viewport:    vp,
		messages:    []message{},
		sender:      user,
		ollamaModel: "codellama:7b",
		styles:      styles,
	}
}

func (m *model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tacmd tea.Cmd
		vpcmd tea.Cmd
		cmd  tea.Cmd
	)

	m.textarea, tacmd = m.textarea.Update(msg)
	m.viewport, vpcmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			userInput := m.textarea.Value()
			if strings.HasPrefix(userInput, "/model ") {
				parts := strings.SplitN(userInput, " ", 2)
				if len(parts) == 2 && parts[1] != "" {
					m.ollamaModel = parts[1]
					m.messages = append(m.messages, message{sender: assist, content: fmt.Sprintf("LLM model changed to %s", m.ollamaModel)})
				} else {
					m.messages = append(m.messages, message{sender: assist, content: "Invalid /model command. Usage: /model <model_name>"})
				}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
			m.messages = append(m.messages, message{sender: user, content: userInput})
			m.messages = append(m.messages, message{sender: assist, content: ""})
			m.viewport.SetContent(m.renderMessages())
			m.textarea.Reset()
			m.viewport.GotoBottom()
			cmd = generateStreamCmd(userInput, m.program, m.ollamaModel)
		case tea.KeyCtrlE:
			if m.command != "" {
				m.messages = append(m.messages, message{sender: execSender, content: ""})
				cmd = execCmd(m.command, m.program)
			}
		}

	case ollamaStreamMsg:
		last := len(m.messages) - 1
		m.messages[last].content += string(msg)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case ollamaStreamDoneMsg:
		last := len(m.messages) - 1
		m.command = parseCommand(m.messages[last].content)

	case execMsg:
		last := len(m.messages) - 1
		m.messages[last].content += string(msg) + "\n"
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case execDoneMsg:
		// Execution is done
	}

	return m, tea.Batch(tacmd, vpcmd, cmd)
}

func (m *model) View() string {
	header := m.styles.headerStyle.Render("KubeMage")
	var commandPreview string
	if m.command != "" {
		commandPreview = fmt.Sprintf("\n---\nCommand Preview (Ctrl+E to execute):\n%s\n---", m.command)
	}

	return fmt.Sprintf(
		"%s\n%s\n%s%s",
		header,
		m.viewport.View(),
		m.textarea.View(),
		commandPreview,
	)
}

func (m *model) renderMessages() string {
	var history string
	for _, msg := range m.messages {
		var senderStyle lipgloss.Style
		senderText := msg.sender
		switch msg.sender {
		case user:
			senderStyle = m.styles.userStyle
			senderText = "You"
		case assist:
			senderStyle = m.styles.assistStyle
			senderText = "KubeMage"
		case execSender:
			senderStyle = m.styles.execStyle
			senderText = "Command"
		}
		history += senderStyle.Render(senderText) + ": " + msg.content + "\n"
	}
	return history
}

func parseCommand(response string) string {
	start := strings.Index(response, "```")
	if start == -1 {
		return ""
	}
	start += 3
	end := strings.Index(response[start:], "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(response[start : start+end])
}
