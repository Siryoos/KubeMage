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
		assistStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")),  // Green
		execStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")), // Yellow
		errorStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")), // Red
		viewportStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1),
	}
}

const (
	user             = "User"
	assist           = "Assistant"
	execSender       = "Exec"
	chatHistoryLimit = 10
	waitingMessage   = "⌛ Waiting for Ollama..."
)

type ollamaStreamMsg string
type ollamaStreamDoneMsg struct{}

func generateStreamCmd(history []message, p *tea.Program, model string) tea.Cmd {
	return func() tea.Msg {
		prompt := buildChatPrompt(history)
		ch := make(chan string)
		go GenerateChatStream(prompt, ch, model)

		initialResponse, ok := <-ch
		if !ok {
			return ollamaStreamDoneMsg{}
		}

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
	showHelp    bool
}

func InitialModel(defaultModel string) *model {
	ta := textarea.New()
	ta.Placeholder = "Ask KubeMage a question..."
	ta.Focus()

	vp := viewport.New(80, 20)
	styles := defaultStyles()
	vp.Style = styles.viewportStyle

	modelName := strings.TrimSpace(defaultModel)
	if modelName == "" {
		modelName = defaultModelName
	}

	selectedModel := modelName
	statusMessage := ""
	if resolved, status, err := resolveModel(modelName, true); err != nil {
		statusMessage = fmt.Sprintf("⚠️ %s", err.Error())
		selectedModel = modelName
	} else {
		selectedModel = resolved
		statusMessage = status
	}

	welcome := message{sender: assist, content: "Welcome to KubeMage! Ask for a kubectl/helm action (e.g. 'List pods in default'), then review the suggested command. Press Ctrl+H for help."}

	m := &model{
		textarea:    ta,
		viewport:    vp,
		messages:    []message{welcome},
		sender:      user,
		ollamaModel: selectedModel,
		styles:      styles,
	}

	if strings.TrimSpace(statusMessage) != "" {
		m.messages = append(m.messages, message{sender: assist, content: statusMessage})
		if strings.HasPrefix(statusMessage, "⚠️") {
			m.showHelp = true
		}
	}

	m.viewport.SetContent(m.renderMessages())
	return m
}

func (m *model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tacmd tea.Cmd
		vpcmd tea.Cmd
		cmd   tea.Cmd
	)

	m.textarea, tacmd = m.textarea.Update(msg)
	m.viewport, vpcmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlH:
			m.showHelp = !m.showHelp
			return m, nil
		case tea.KeyEnter:
			userInput := m.textarea.Value()
			if strings.HasPrefix(userInput, "/model ") {
				parts := strings.SplitN(userInput, " ", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
					m.ollamaModel = strings.TrimSpace(parts[1])
					m.messages = append(m.messages, message{sender: assist, content: fmt.Sprintf("LLM model changed to %s", m.ollamaModel)})
				} else {
					m.messages = append(m.messages, message{sender: assist, content: "Invalid /model command. Usage: /model <model_name>"})
				}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
			trimmed := strings.TrimSpace(userInput)
			if trimmed == "" {
				return m, nil
			}
			m.command = ""
			m.messages = append(m.messages, message{sender: user, content: trimmed})
			history := append([]message(nil), m.messages...)
			m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
			m.viewport.SetContent(m.renderMessages())
			m.textarea.Reset()
			m.viewport.GotoBottom()
			cmd = generateStreamCmd(history, m.program, m.ollamaModel)
		case tea.KeyCtrlE:
			if m.command != "" {
				m.messages = append(m.messages, message{sender: execSender, content: fmt.Sprintf("$ %s\n", m.command)})
				cmd = execCmd(m.command, m.program)
			}
		case tea.KeyCtrlK:
			m.command = ""
		}

	case ollamaStreamMsg:
		last := len(m.messages) - 1
		if m.messages[last].content == waitingMessage {
			m.messages[last].content = ""
		}
		m.messages[last].content += string(msg)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case ollamaStreamDoneMsg:
		last := len(m.messages) - 1
		assistantReply := m.messages[last].content
		if strings.HasPrefix(assistantReply, "Error") {
			m.command = ""
			break
		}
		m.command = parseCommand(assistantReply)
		if m.command == "" {
			if strings.TrimSpace(assistantReply) == "" {
				m.messages[last].content = "I didn’t receive any text from the model."
			}
			m.messages = append(m.messages, message{sender: assist, content: "I could not find a runnable command in that response."})
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}

	case execMsg:
		last := len(m.messages) - 1
		m.messages[last].content += string(msg) + "\n"
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case execDoneMsg:
		m.command = ""
	}

	return m, tea.Batch(tacmd, vpcmd, cmd)
}

func (m *model) View() string {
	header := m.styles.headerStyle.Render(fmt.Sprintf("KubeMage [%s]", m.ollamaModel))
	var commandPreview string
	if m.command != "" {
		commandPreview = fmt.Sprintf("\n---\nCommand Preview (Ctrl+E to execute, Ctrl+K to clear):\n%s\n---", m.command)
	}

	help := ""
	if m.showHelp {
		help = "\n---\nShortcuts: Enter to ask • Ctrl+E run preview • Ctrl+K clear preview • Ctrl+H toggle help • Type /model <name> to switch models. Ensure 'ollama serve' is running or set OLLAMA_HOST."
	}

	return fmt.Sprintf(
		"%s\n%s\n%s%s%s",
		header,
		m.viewport.View(),
		m.textarea.View(),
		help,
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
	segment := strings.TrimSpace(response[start : start+end])
	if newline := strings.Index(segment, "\n"); newline != -1 {
		languageLine := strings.TrimSpace(segment[:newline])
		if languageLine != "" && !strings.ContainsAny(languageLine, " \t") {
			segment = strings.TrimSpace(segment[newline+1:])
		}
	}
	return segment
}

func buildChatPrompt(history []message) string {
	if len(history) == 0 {
		return ""
	}

	start := 0
	if len(history) > chatHistoryLimit {
		start = len(history) - chatHistoryLimit
	}
	trimmed := history[start:]

	var sb strings.Builder
	for _, msg := range trimmed {
		content := strings.TrimSpace(msg.content)
		if content == "" {
			continue
		}

		label := "User"
		switch msg.sender {
		case assist:
			label = "Assistant"
		case execSender:
			label = "Command Output"
		}

		sb.WriteString(label)
		sb.WriteString(": ")
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Assistant:")
	return sb.String()
}
