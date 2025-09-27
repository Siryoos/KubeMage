package main

import (
	"fmt"
	"os"
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
	systemStyle   lipgloss.Style
	yamlKeyStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	viewportStyle lipgloss.Style
	headerStyle   lipgloss.Style
}

func defaultStyles() styles {
	return styles{
		userStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),  // Blue
		assistStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")),  // Green
		execStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")), // Yellow
		systemStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")), // Gray
		yamlKeyStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("178")),             // Orange
		errorStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")), // Red
		viewportStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("57")).Padding(0, 1),
	}
}

const (
	user           = "User"
	assist         = "Assistant"
	execSender     = "Exec"
	systemSender   = "System"
	waitingMessage = "⌛ Waiting for Ollama..."
)

type ollamaStreamMsg string
type ollamaStreamDoneMsg struct{}

func generateStreamCmd(m *model, history []message) tea.Cmd {
	return func() tea.Msg {
		prompt := m.buildChatPrompt(history)
		ch := make(chan string)
		systemPrompt := chatAssistantSystemPrompt
		if m.agentMode {
			systemPrompt = agentSystemPrompt
		}
		go GenerateChatStream(prompt, ch, m.ollamaModel, systemPrompt)

		initialResponse, ok := <-ch
		if !ok {
			return ollamaStreamDoneMsg{}
		}

		go func() {
			for response := range ch {
				m.program.Send(ollamaStreamMsg(response))
			}
			m.program.Send(ollamaStreamDoneMsg{})
		}()

		return ollamaStreamMsg(initialResponse)
	}
}

type message struct {
	sender  string
	content string
}

type model struct {
	program               *tea.Program
	viewport              viewport.Model
	textarea              textarea.Model
	messages              []message
	sender                string
	command               string
	stdoutContent         map[string]string
	stderrContent         map[string]string
	ollamaModel           string
	styles                styles
	showHelp              bool
	agentMode             bool
	agentState            string // "", "thinking", "acting"
	awaitingSecondConfirm *PreExecPlan
	config                *Config
	metrics               *Metrics
	dumpMetrics           bool
}

func InitialModel(defaultModel string, cfg *Config, dumpMetrics bool) *model {
	ta := textarea.New()
	ta.Placeholder = "Ask KubeMage a question..."
	ta.Focus()

	vp := viewport.New(80, 20)
	styles := defaultStyles()
	vp.Style = styles.viewportStyle

	modelName := strings.TrimSpace(defaultModel)
	if modelName == "" {
		modelName = cfg.Model
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
		textarea:      ta,
		viewport:      vp,
		messages:      []message{welcome},
		sender:        user,
		ollamaModel:   selectedModel,
		styles:        styles,
		stdoutContent: make(map[string]string),
		stderrContent: make(map[string]string),
		config:        cfg,
		metrics:       &Metrics{},
		dumpMetrics:   dumpMetrics,
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

func (m *model) DumpMetrics() {
	if m.dumpMetrics {
		m.metrics.Dump()
	}
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
			m.DumpMetrics()
			return m, tea.Quit
		case tea.KeyCtrlH:
			m.showHelp = !m.showHelp
			return m, nil
		case tea.KeyEnter:
			userInput := m.textarea.Value()
			if strings.HasPrefix(userInput, "/agent") {
				m.agentMode = !m.agentMode
				if m.agentMode {
					m.agentState = "thinking"
					m.messages = append(m.messages, message{sender: systemSender, content: "Agent mode activated."})
				} else {
					m.agentState = ""
					m.messages = append(m.messages, message{sender: systemSender, content: "Agent mode deactivated."})
				}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/save ") {
				parts := strings.Fields(userInput)
				if len(parts) >= 2 {
					filename := parts[1]
					var lastAssistantMsg string
					for i := len(m.messages) - 1; i >= 0; i-- {
						if m.messages[i].sender == assist {
							lastAssistantMsg = m.messages[i].content
							break
						}
					}

					if lastAssistantMsg != "" {
						// Create out directory if it doesn't exist
						if _, err := os.Stat("out"); os.IsNotExist(err) {
							os.Mkdir("out", 0755)
						}

						err := os.WriteFile("out/"+filename, []byte(lastAssistantMsg), 0644)
						if err != nil {
							m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error saving file: %v", err)})
						} else {
							m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Saved to out/%s", filename)})
							// Suggest next command
							if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
								suggestedCmd := fmt.Sprintf("kubectl apply --dry-run=client -f out/%s", filename)
								m.messages = append(m.messages, message{sender: assist, content: "You can now run a dry-run apply with the following command:"})
								m.command = suggestedCmd
							}
						}
					} else {
						m.messages = append(m.messages, message{sender: systemSender, content: "No assistant message to save."})
					}
				} else {
					m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /save <filename>"})
				}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/gen-deploy") {
				parts := strings.Fields(userInput)
				name := "example"
				if len(parts) > 1 {
					name = parts[1]
				}
				prompt := GetManifestGenerationPrompt("Deployment", name)
				m.messages = append(m.messages, message{sender: user, content: prompt})
				history := append([]message(nil), m.messages...)
				m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
				m.viewport.SetContent(m.renderMessages())
				m.textarea.Reset()
				m.viewport.GotoBottom()
				cmd = generateStreamCmd(m, history)
				return m, tea.Batch(tacmd, vpcmd, cmd)
			}
			if strings.HasPrefix(userInput, "/gen-helm") {
				parts := strings.Fields(userInput)
				name := "example"
				if len(parts) > 1 {
					name = parts[1]
				}
				prompt := GetHelmGenerationPrompt(name)
				m.messages = append(m.messages, message{sender: user, content: prompt})
				history := append([]message(nil), m.messages...)
				m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
				m.viewport.SetContent(m.renderMessages())
				m.textarea.Reset()
				m.viewport.GotoBottom()
				cmd = generateStreamCmd(m, history)
				return m, tea.Batch(tacmd, vpcmd, cmd)
			}
			if strings.HasPrefix(userInput, "/diag-pod ") {
				parts := strings.Fields(userInput)
				if len(parts) >= 2 {
					ns, _ := GetCurrentNamespace()
					pod := parts[1]
					results, _ := DiagnosePodNotReady(pod, ns)
					for _, r := range results {
						m.messages = append(m.messages, message{sender: execSender, content: "$ " + r.Command})
						m.messages = append(m.messages, message{sender: systemSender, content: r.Output})
						for _, note := range r.Notes {
							m.messages = append(m.messages, message{sender: systemSender, content: "Note: " + note})
						}
					}
					m.viewport.SetContent(m.renderMessages())
					m.viewport.GotoBottom()
					return m, nil
				}
			}
			if userInput == "/model list" {
				models, err := ListModels()
				if err != nil {
					m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error getting models: %v", err)})
				} else {
					m.messages = append(m.messages, message{sender: assist, content: "Available models:\n" + strings.Join(models, "\n")})
				}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
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
			cmd = generateStreamCmd(m, history)
		case tea.KeyCtrlE:
			if m.awaitingSecondConfirm != nil {
				// Execute original for real
			realCmd := m.awaitingSecondConfirm.Original
			m.messages = append(m.messages, message{sender: execSender, content: "$ " + realCmd})
			m.viewport.SetContent(m.renderMessages())
				m.metrics.ExecutedCommands++
				cmd = execCmd(realCmd, m.program)
				m.awaitingSecondConfirm = nil
				break
			}

			if m.command != "" {
				plan := BuildPreExecPlan(m.command)
				m.metrics.ValidatedCommands++
				m.messages = append(m.messages, message{sender: systemSender, content: plan.HumanPreview()})
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()

				var cmds []tea.Cmd
				for _, ch := range plan.Checks {
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + ch.Cmd})
					m.viewport.SetContent(m.renderMessages())
						m.metrics.ExecutedCommands++
					cmds = append(cmds, execCmd(ch.Cmd, m.program))
				}

				if plan.FirstRunCommand != "" {
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + plan.FirstRunCommand})
					m.viewport.SetContent(m.renderMessages())
						m.metrics.ExecutedCommands++
					cmds = append(cmds, execCmd(plan.FirstRunCommand, m.program))
				}

				if plan.RequireSecondConfirm {
					m.messages = append(m.messages, message{sender: systemSender, content: "Press Ctrl+E again to APPLY for real, or edit the command first."})
					m.awaitingSecondConfirm = &plan
				}
				cmd = tea.Sequence(cmds...)
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

			if m.agentMode && m.agentState == "thinking" {
				if action := parseAction(assistantReply); action != "" {
					m.metrics.TotalCommands++
					if isActionWhitelisted(action) {
						m.agentState = "acting"
						m.messages = append(m.messages, message{sender: execSender, content: "$ " + action})
						m.metrics.ExecutedCommands++
						cmd = execCmd(action, m.program)
					} else {
						m.messages = append(m.messages, message{sender: systemSender, content: "Action not allowed."})
						m.agentState = ""
						m.agentMode = false
					}
				} else if finalAnswer := parseFinalAnswer(assistantReply); finalAnswer != "" {
					m.messages[last].content = finalAnswer
					m.agentState = ""
					m.agentMode = false
				}
		} else if !m.agentMode {
			if strings.HasPrefix(assistantReply, "Error") {
				m.command = ""
				break
			}
			m.command = parseCommand(assistantReply)
			if m.command != "" {
				m.metrics.TotalCommands++
			}
			if m.command == "" {
				if strings.TrimSpace(assistantReply) == "" {
					m.messages[last].content = "I didn’t receive any text from the model."
				}
				m.messages = append(m.messages, message{sender: assist, content: "I could not find a runnable command in that response."})
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
			}
		}

	case stdoutMsg:
		m.stdoutContent[msg.cmd] += msg.out + "\n"
		// Find the message for this command and append output
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].sender == execSender && strings.HasSuffix(m.messages[i].content, msg.cmd) {
				m.messages[i].content += "\n" + msg.out
				break
			}
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case stderrMsg:
		m.stderrContent[msg.cmd] += msg.out + "\n"
		m.messages = append(m.messages, message{sender: systemSender, content: "stderr: " + msg.out})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case execDoneMsg:
		if m.agentMode && m.agentState == "acting" {
			observation := ""
			if msg.err != nil {
				observation = fmt.Sprintf("Error: %v", msg.err)
				if stderr, ok := m.stderrContent[msg.cmd]; ok && stderr != "" {
					observation += "\n" + stderr
				}
			} else {
				if stdout, ok := m.stdoutContent[msg.cmd]; ok && stdout != "" {
					observation = stdout
				} else {
					observation = "Command executed successfully with no output."
				}
			}
			
			m.messages = append(m.messages, message{sender: user, content: "Observation: " + observation})
			history := append([]message(nil), m.messages...)
			m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
			m.viewport.SetContent(m.renderMessages())
			m.textarea.Reset()
			m.viewport.GotoBottom()
			cmd = generateStreamCmd(m, history)
			m.agentState = "thinking"

		} else if msg.err != nil {
			// Command failed
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Command '%s' failed: %v", msg.cmd, msg.err)})

			if stderr, ok := m.stderrContent[msg.cmd]; ok && stderr != "" {
				// There was stderr, trigger self-correction
				correctionPrompt := fmt.Sprintf("The command `%s` failed with the error:\n```\n%s\n```\nCan you fix it?", msg.cmd, stderr)
				
				m.messages = append(m.messages, message{sender: user, content: correctionPrompt})
				history := append([]message(nil), m.messages...)
				m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
				m.viewport.SetContent(m.renderMessages())
				m.textarea.Reset()
				m.viewport.GotoBottom()
				cmd = generateStreamCmd(m, history)
			}
		}
		m.command = ""
		delete(m.stdoutContent, msg.cmd)
		delete(m.stderrContent, msg.cmd)
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
	help = "\n---\nShortcuts: Enter to ask • Ctrl+E run preview • Ctrl+K clear preview • Ctrl+H toggle help • Type /model <name> to switch models.\n\nEnsure 'ollama serve' is running locally or set OLLAMA_HOST.\nFor remote Ollama, use SSH port forwarding:\nssh -f -N -L 11434:localhost:11434 user@host"
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
		content := msg.content
		switch msg.sender {
		case user:
			senderStyle = m.styles.userStyle
			senderText = "You"
		case assist:
			senderStyle = m.styles.assistStyle
			senderText = "KubeMage"
			if strings.Contains(content, "apiVersion:") {
				content = m.highlightYAML(content)
			}
		case execSender:
			senderStyle = m.styles.execStyle
			senderText = "Command"
		case systemSender:
			senderStyle = m.styles.systemStyle
			senderText = "System"
		}
		if len(content) > m.config.TruncationSize {
			content = content[:m.config.TruncationSize] + "\n(...truncated...)"
		}
		history += senderStyle.Render(senderText) + ": " + content + "\n"
	}
	return history
}

func (m *model) highlightYAML(content string) string {
	keywords := []string{"apiVersion:", "kind:", "metadata:", "spec:", "labels:", "selector:", "template:", "containers:", "image:", "ports:", "name:", "replicas:", "app:"}
	highlightedContent := content
	for _, keyword := range keywords {
			highlightedContent = strings.ReplaceAll(highlightedContent, keyword, m.styles.yamlKeyStyle.Render(keyword))
	}
	return highlightedContent
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

func parseFinalAnswer(response string) string {
	start := strings.Index(response, "Final:")
	if start == -1 {
		return ""
	}
	start += len("Final:")
	return strings.TrimSpace(response[start:])
}

var whitelistedActions = []string{"kubectl get", "kubectl describe", "kubectl logs", "kubectl events"}

func isActionWhitelisted(action string) bool {
	for _, prefix := range whitelistedActions {
		if strings.HasPrefix(action, prefix) {
			return true
		}
	}
	return false
}

func parseAction(response string) string {
	start := strings.Index(response, "Action:")
	if start == -1 {
		return ""
	}
	start += len("Action:")
	segment := strings.TrimSpace(response[start:])
	// The action might be in a code block
	if strings.HasPrefix(segment, "```") {
		end := strings.Index(segment[3:], "```")
		if end != -1 {
			segment = segment[3 : 3+end]
		}
	}
	return strings.TrimSpace(segment)
}

func (m *model) buildChatPrompt(history []message) string {
	if len(history) == 0 {
		return ""
	}

	start := 0
	if len(history) > m.config.ChatHistoryLength {
		start = len(history) - m.config.ChatHistoryLength
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