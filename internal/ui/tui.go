package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/siryoos/kubemage/internal/config"
	"github.com/siryoos/kubemage/internal/engine"
	"github.com/siryoos/kubemage/internal/engine/validator"
	"github.com/siryoos/kubemage/internal/execx"
	"github.com/siryoos/kubemage/internal/metrics"
)

type styles struct {
	// Core message styles
	userStyle     lipgloss.Style
	assistStyle   lipgloss.Style
	execStyle     lipgloss.Style
	systemStyle   lipgloss.Style
	yamlKeyStyle  lipgloss.Style
	errorStyle    lipgloss.Style

	// Panel and layout styles
	viewportStyle lipgloss.Style
	diffStyle     lipgloss.Style
	headerStyle   lipgloss.Style
	statusStyle   lipgloss.Style
	hintBoxStyle  lipgloss.Style
	hintKeyStyle  lipgloss.Style
	hintDescStyle lipgloss.Style
	inputWrapper  lipgloss.Style
	footerStyle   lipgloss.Style

	// Context and status styles
	contextStyle  lipgloss.Style
	contextAlert  lipgloss.Style
	riskLowStyle  lipgloss.Style
	riskMedStyle  lipgloss.Style
	riskHighStyle lipgloss.Style
	riskCritStyle lipgloss.Style

	// Intelligence UI styles
	suggestionStyle     lipgloss.Style
	insightStyle        lipgloss.Style
	quickActionStyle    lipgloss.Style
	confidenceStyle     lipgloss.Style
	healthyIndicator    lipgloss.Style
	warningIndicator    lipgloss.Style
	errorIndicator      lipgloss.Style

	// Interactive elements
	hotKeyStyle         lipgloss.Style
	clickableStyle      lipgloss.Style
	selectedStyle       lipgloss.Style
	badgeStyle          lipgloss.Style
	progressStyle       lipgloss.Style
	spacerStyle         lipgloss.Style
}

type layoutMode int

const (
	layoutThreePane layoutMode = iota
	layoutChatOnly
	layoutVerticalSplit
	layoutHorizontalSplit
)

type rightPaneMode int

const (
	rightPaneText rightPaneMode = iota
	rightPaneDiff
)

func (l layoutMode) next() layoutMode {
	switch l {
	case layoutThreePane:
		return layoutVerticalSplit
	case layoutVerticalSplit:
		return layoutHorizontalSplit
	case layoutHorizontalSplit:
		return layoutChatOnly
	default:
		return layoutThreePane
	}
}

const (
	contextRefreshInterval = 15 * time.Second
	clockTickInterval      = time.Second
	maxLiveTokens          = 9999
)

func defaultStyles() styles {
	return styles{
		// Core message styles with improved colors
		userStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4FF")),  // Bright cyan
		assistStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF88")),  // Bright green
		execStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700")), // Gold
		systemStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8A8A8A")), // Gray
		yamlKeyStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C42")),            // Orange
		errorStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4757")), // Red

		// Enhanced panel styles with modern borders
		viewportStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5F87AF")).Padding(0, 1),
		diffStyle:     lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#AF87FF")).Padding(0, 1),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#2E3440")).Padding(0, 2),
		statusStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D8DEE9")).Padding(0, 1),
		hintBoxStyle:  lipgloss.NewStyle().MarginTop(1).Padding(1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4C566A")).Align(lipgloss.Left),
		hintKeyStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")).Bold(true),
		hintDescStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#D8DEE9")),
		inputWrapper:  lipgloss.NewStyle().MarginTop(1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5E81AC")).Padding(0, 1),
		footerStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")).MarginTop(1),

		// Context awareness styles
		contextStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")).Background(lipgloss.Color("#3B4252")).Padding(0, 1),
		contextAlert:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")).Background(lipgloss.Color("#BF616A")).Bold(true).Padding(0, 1),

		// Risk level indicators
		riskLowStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Bold(true),
		riskMedStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Bold(true),
		riskHighStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#D08770")).Bold(true),
		riskCritStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")).Bold(true).Blink(true),

		// Intelligence UI styles
		suggestionStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")).Background(lipgloss.Color("#434C5E")).Padding(0, 1).Margin(0, 1),
		insightStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5E81AC")).Padding(1),
		quickActionStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")).Background(lipgloss.Color("#5E81AC")).Padding(0, 1).Margin(0, 1),
		confidenceStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Italic(true),
		healthyIndicator:    lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		warningIndicator:    lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
		errorIndicator:      lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),

		// Interactive elements
		hotKeyStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true).Underline(true),
		clickableStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")).Underline(true),
		selectedStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")).Background(lipgloss.Color("#5E81AC")),
		badgeStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3440")).Background(lipgloss.Color("#D08770")).Padding(0, 1),
		progressStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		spacerStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
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

type commandHint struct {
	Trigger     string
	Description string
}

var commandPalette = []commandHint{
	{"/help", "Toggle comprehensive help with new features"},
	{"/model list", "List available Ollama models"},
	{"/model set chat <name>", "Switch chat assistant model"},
	{"/model set generation <name>", "Switch generation/diff model"},
	{"/edit-yaml <path> <instruction>", "Generate a diff for a manifest"},
	{"/edit-values <path> <instruction>", "Generate a diff for Helm values"},
	{"/gen-deploy <name> --image <img>", "Draft a deployment manifest"},
	{"/gen-helm <chart> [flags]", "Generate a Helm chart skeleton"},
	{"/diag-pod <name>", "Run intelligent pod diagnostics"},
	{"/agent", "Toggle ReAct agent mode"},
	{"/ctx", "Show current cluster context"},
	{"/ns set <namespace>", "Switch active namespace"},
	{"/metrics", "Show comprehensive session metrics"},
	{"/resolve [note]", "Mark the current task as resolved"},
}

type contextSummaryMsg struct {
	summary *engine.KubeContextSummary
	err     error
}

type clockTickMsg time.Time

type commandTrackedMsg struct {
	command string
}

type asyncIntelligenceResultMsg struct {
	result engine.IntelligenceResult
}

func generateStreamCmd(m *model, history []message, modelName string) tea.Cmd {
	return func() tea.Msg {
		prompt := m.buildChatPrompt(history)
		ch := make(chan string)
		systemPrompt := chatAssistantSystemPrompt
		if m.agentMode {
			systemPrompt = agentSystemPrompt
		}
		go GenerateChatStream(prompt, ch, modelName, systemPrompt)

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
	chatViewport          viewport.Model
	previewViewport       viewport.Model
	outputViewport        viewport.Model
	textarea              textarea.Model
	messages              []message
	sender                string
	command               string
	activeCommand         string
	stdoutContent         map[string]string
	stderrContent         map[string]string
	ollamaModel           string
	generationModel       string
	styles                styles
	showHelp              bool
	agentMode             bool
	agentState            string // "", "thinking", "acting"
	awaitingSecondConfirm *validator.PreExecPlan
	awaitingTypedConfirm  *validator.PreExecPlan
	currentPlan           *validator.PreExecPlan
	previewCheckResults   map[string]execx.PreviewCheckDoneMsg
	config                *config.AppConfig
	metrics               *metrics.SessionMetrics
	dumpMetrics           bool
	metricsFlushed        bool
	pendingDiff           *DiffSession
	pendingGeneration     *GenerationSession
	windowWidth           int
	windowHeight          int
	layout                layoutMode
	rightTopMode          rightPaneMode
	ctxName               string
	namespace             string
	rbacUser              string
	liveTokens            int
	lastFooterUpdate      time.Time

	// Enhanced intelligence features
	intelligentUI         *IntelligentUI
	currentContext        *KubeContextSummary
	lastIntelligenceUpdate time.Time
	showSuggestions       bool
	showInsights          bool
	selectedSuggestion    int
	clusterHealth         string
	riskLevel             RiskLevel

	// Smart refresh tracking
	lastContextHash       string
	lastUserInput         time.Time
	recentCommands        []string
	maxRecentCommands     int
	userTypingTimeout     time.Duration

	// Async intelligence tracking
	pendingIntelligenceWork map[string]bool
	asyncIntelligenceEnabled bool

	// Streaming intelligence
	streamingManager *engine.StreamingIntelligenceManager
	intelligenceSubscriber *StreamSubscriber
}

func InitialModel(defaultModel string, cfg *Config, dumpMetrics bool) *model {
	ta := textarea.New()
	ta.Placeholder = "Ask KubeMage a question..."
	ta.Prompt = "❯ "
	ta.SetWidth(78)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Focus()

	chatVP := viewport.New(80, 20)
	previewVP := viewport.New(40, 10)
	outputVP := viewport.New(40, 10)
	styles := defaultStyles()
	chatVP.Style = styles.viewportStyle
	previewVP.Style = styles.viewportStyle
	outputVP.Style = styles.viewportStyle

	modelName := strings.TrimSpace(defaultModel)
	if modelName == "" {
		modelName = strings.TrimSpace(cfg.Models.Chat)
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

	rbacUser := strings.TrimSpace(os.Getenv("USER"))
	if rbacUser == "" {
		rbacUser = "(user)"
	}

	m := &model{
		textarea:            ta,
		chatViewport:        chatVP,
		previewViewport:     previewVP,
		outputViewport:      outputVP,
		messages:            []message{welcome},
		sender:              user,
		ollamaModel:         selectedModel,
		generationModel:     strings.TrimSpace(cfg.Models.Generation),
		styles:              styles,
		stdoutContent:       make(map[string]string),
		stderrContent:       make(map[string]string),
		previewCheckResults: make(map[string]previewCheckDoneMsg),
		config:              cfg,
		metrics:             NewSessionMetrics(),
		dumpMetrics:         dumpMetrics,
		metricsFlushed:      false,
		layout:              layoutThreePane,
		rightTopMode:        rightPaneText,
		ctxName:             "(ctx)",
		namespace:           "(ns)",
		rbacUser:            rbacUser,

		// Initialize intelligence features
		intelligentUI:         NewIntelligentUI(),
		showSuggestions:       true,
		showInsights:          true,
		selectedSuggestion:    -1,
		clusterHealth:         "unknown",
		riskLevel:             RiskLevel{Level: "low", Factors: []string{}, Reversible: true},
		lastIntelligenceUpdate: time.Time{},

		// Initialize smart refresh tracking
		lastContextHash:       "",
		lastUserInput:         time.Time{},
		recentCommands:        make([]string, 0),
		maxRecentCommands:     10,
		userTypingTimeout:     3 * time.Second,

		// Initialize async intelligence tracking
		pendingIntelligenceWork:  make(map[string]bool),
		asyncIntelligenceEnabled: true,
	}

	// Initialize streaming intelligence
	m.streamingManager = NewStreamingIntelligenceManager(nil) // Program will be set later
	filters := []UpdateFilter{
		{Type: UpdateTypeRiskChange, MinPriority: 5},
		{Type: UpdateTypeNewSuggestion, MinPriority: 3},
		{Type: UpdateTypeContextChange, MinPriority: 7},
		{Type: UpdateTypeHealthChange, MinPriority: 6},
	}
	m.intelligenceSubscriber = m.streamingManager.Subscribe("main_ui", filters, 50)

	// Initialize predictive intelligence if smart cache is available
	if SmartCache != nil {
		InitializePredictiveIntelligence(SmartCache, m.streamingManager)
	}

	// Initialize model router for intelligent model selection
	if SmartCache != nil {
		InitializeModelRouter(SmartCache)
	}

	// Initialize adaptive UI manager
	if m.streamingManager != nil {
		InitializeAdaptiveUI(m.streamingManager)
	}

	// Initialize performance optimizer
	InitializePerformanceOptimizer()

	// Initialize intelligent command generator
	if ModelRouter != nil && SmartCache != nil {
		InitializeIntelligentCommandGenerator(ModelRouter, SmartCache)
	}

	// Initialize real-time performance monitor
	InitializeRealTimePerformanceMonitor()

	if strings.TrimSpace(m.generationModel) == "" {
		m.generationModel = m.ollamaModel
	}

	if strings.TrimSpace(statusMessage) != "" {
		m.messages = append(m.messages, message{sender: assist, content: statusMessage})
		if strings.HasPrefix(statusMessage, "⚠️") {
			m.showHelp = true
		}
	}

	m.chatViewport.SetContent(m.renderMessages())
	m.refreshPreviewPane()
	m.refreshOutputPane()

	// Initialize async intelligence processing
	if m.asyncIntelligenceEnabled {
		Intelligence.InitializeAsyncProcessor(m.program)
	}

	return m
}

func (m *model) updateLayout() {
	if m.windowWidth == 0 || m.windowHeight == 0 {
		return
	}

	// Smart responsive layout adjustments
	contentWidth := m.windowWidth - 6
	if contentWidth < 60 {
		contentWidth = 60
	}

	// Intelligent height calculation considering intelligence panels
	baseHeight := m.windowHeight - 12
	intelligencePanelHeight := m.calculateIntelligencePanelHeight()
	contentHeight := baseHeight - intelligencePanelHeight

	if contentHeight < 15 {
		contentHeight = 15
		// If too cramped, reduce intelligence panel
		intelligencePanelHeight = baseHeight - 15
		if intelligencePanelHeight < 0 {
			intelligencePanelHeight = 0
		}
	}

	// Adaptive layout based on screen size and content
	isSmallScreen := m.windowWidth < 120 || m.windowHeight < 40
	isMediumScreen := m.windowWidth < 160 || m.windowHeight < 50

	chatWidth := contentWidth
	chatHeight := contentHeight
	previewWidth := contentWidth
	previewHeight := contentHeight / 2
	outputWidth := contentWidth
	outputHeight := contentHeight - previewHeight

	switch m.layout {
	case layoutThreePane:
		if isSmallScreen {
			// Force vertical layout on small screens
			m.layout = layoutVerticalSplit
			m.updateLayout()
			return
		}

		// Intelligent width allocation based on content
		chatRatio := 0.55
		if m.intelligentUI != nil && len(m.intelligentUI.suggestions) > 0 {
			chatRatio = 0.5 // More space for preview when suggestions exist
		}

		chatWidth = int(float64(contentWidth) * chatRatio)
		if chatWidth < 40 {
			chatWidth = 40
		}
		previewWidth = contentWidth - chatWidth - 1
		if previewWidth < 30 {
			previewWidth = 30
		}
		outputWidth = previewWidth

		// Dynamic height allocation based on content importance
		previewRatio := 0.5
		if m.rightTopMode == rightPaneDiff {
			previewRatio = 0.6 // More space for diffs
		} else if m.activeCommand != "" {
			previewRatio = 0.4 // More space for output when command running
		}

		previewHeight = int(float64(contentHeight) * previewRatio)
		if previewHeight < 8 {
			previewHeight = 8
		}
		outputHeight = contentHeight - previewHeight - 1
		if outputHeight < 6 {
			outputHeight = 6
		}
		chatHeight = contentHeight

	case layoutVerticalSplit:
		chatRatio := 0.6
		if isMediumScreen {
			chatRatio = 0.7 // More chat space on medium screens
		}

		chatWidth = int(float64(contentWidth) * chatRatio)
		if chatWidth < 45 {
			chatWidth = 45
		}
		previewWidth = contentWidth - chatWidth - 1
		outputWidth = previewWidth
		previewHeight = contentHeight
		outputHeight = contentHeight
		chatHeight = contentHeight

	case layoutHorizontalSplit:
		chatWidth = contentWidth
		chatRatio := 0.55
		if m.intelligentUI != nil && m.intelligentUI.IsHighRisk() {
			chatRatio = 0.5 // More space for preview in high-risk situations
		}

		chatHeight = int(float64(contentHeight) * chatRatio)
		if chatHeight < 10 {
			chatHeight = 10
		}
		previewWidth = contentWidth
		previewHeight = contentHeight - chatHeight - 1
		if previewHeight < 6 {
			previewHeight = 6
		}
		outputWidth = previewWidth
		outputHeight = previewHeight

	case layoutChatOnly:
		chatWidth = contentWidth
		chatHeight = contentHeight
		previewWidth = 0
		previewHeight = 0
		outputWidth = 0
		outputHeight = 0
	}

	// Apply layout with smooth transitions
	yOffset := m.chatViewport.YOffset
	m.chatViewport.Width = chatWidth
	m.chatViewport.Height = chatHeight
	m.chatViewport.SetContent(m.renderMessages())
	if yOffset > m.chatViewport.YOffset {
		m.chatViewport.YOffset = yOffset
	}

	if previewWidth > 0 && previewHeight > 0 {
		m.previewViewport.Width = previewWidth
		m.previewViewport.Height = previewHeight
	} else {
		m.previewViewport.Width = 0
		m.previewViewport.Height = 0
	}

	if outputWidth > 0 && outputHeight > 0 {
		m.outputViewport.Width = outputWidth
		m.outputViewport.Height = outputHeight
	} else {
		m.outputViewport.Width = 0
		m.outputViewport.Height = 0
	}

	// Smart textarea width
	textareaWidth := chatWidth
	if m.layout == layoutThreePane && previewWidth > 0 {
		textareaWidth = contentWidth
	}
	if textareaWidth < 40 {
		textareaWidth = 40
	}

	// Adaptive textarea height based on context
	textareaHeight := 3
	if m.intelligentUI != nil && m.intelligentUI.IsHighRisk() {
		textareaHeight = 2 // Smaller input area in high-risk situations
	} else if isSmallScreen {
		textareaHeight = 2
	}

	m.textarea.SetWidth(textareaWidth)
	m.textarea.SetHeight(textareaHeight)
}

// calculateIntelligencePanelHeight estimates space needed for intelligence components
func (m *model) calculateIntelligencePanelHeight() int {
	if m.intelligentUI == nil {
		return 0
	}

	height := 0

	// Suggestions panel
	if m.showSuggestions && len(m.intelligentUI.suggestions) > 0 {
		height += 4 + len(m.intelligentUI.suggestions) // Header + suggestions
	}

	// Insights panel
	if m.showInsights && len(m.intelligentUI.insights) > 0 {
		height += 3 + (len(m.intelligentUI.insights) * 2) // Header + insights with spacing
	}

	// Risk warning
	if m.intelligentUI.IsHighRisk() {
		height += 2
	}

	// Panel borders and margins
	if height > 0 {
		height += 4 // Border + margins
	}

	return height
}

func (m *model) DumpMetrics() {
	if m.dumpMetrics && !m.metricsFlushed {
		m.metrics.Dump()
		m.metricsFlushed = true
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, requestContextSummary(), scheduleClockTick())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tacmd      tea.Cmd
		chatcmd    tea.Cmd
		prevcmd    tea.Cmd
		outcmd     tea.Cmd
		cmd        tea.Cmd
		contextCmd tea.Cmd
		tickCmd    tea.Cmd
	)

	m.textarea, tacmd = m.textarea.Update(msg)
	m.chatViewport, chatcmd = m.chatViewport.Update(msg)
	m.previewViewport, prevcmd = m.previewViewport.Update(msg)
	m.outputViewport, outcmd = m.outputViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.updateLayout()
		return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd)
	case contextSummaryMsg:
		if msg.summary != nil && msg.err == nil {
			m.ctxName = msg.summary.Context
			m.namespace = msg.summary.Namespace
		}
		contextCmd = scheduleContextRefresh()
	case clockTickMsg:
		m.lastFooterUpdate = time.Time(msg)
		tickCmd = scheduleClockTick()
	case commandTrackedMsg:
		m.trackCommand(msg.command)
	case asyncIntelligenceResultMsg:
		m.handleAsyncIntelligenceResult(msg.result)
	case streamingIntelligenceUpdateMsg:
		m.handleStreamingIntelligenceUpdate(msg.update)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyF1:
			// Handle F1 hotkey for first AI suggestion
			if m.intelligentUI != nil {
				if suggestion := m.intelligentUI.HandleHotkey("F1"); suggestion != nil {
					m.executeAISuggestion(suggestion)
				}
			}
			return m, nil
		case tea.KeyF2:
			// F2 for layout switching and second AI suggestion
			if m.intelligentUI != nil && len(m.intelligentUI.suggestions) > 1 {
				if suggestion := m.intelligentUI.HandleHotkey("F2"); suggestion != nil {
					m.executeAISuggestion(suggestion)
					return m, nil
				}
			}
			m.layout = m.layout.next()
			m.updateLayout()
			return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd)
		case tea.KeyF3:
			// Handle F3 hotkey for third AI suggestion
			if m.intelligentUI != nil {
				if suggestion := m.intelligentUI.HandleHotkey("F3"); suggestion != nil {
					m.executeAISuggestion(suggestion)
				}
			}
			return m, nil
		case tea.KeyF4:
			// Toggle suggestions panel
			m.showSuggestions = !m.showSuggestions
			return m, nil
		case tea.KeyF5:
			// Toggle insights panel
			m.showInsights = !m.showInsights
			return m, nil
		case tea.KeyF12:
			// Force refresh intelligence
			m.refreshIntelligence()
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			m.DumpMetrics()
			return m, tea.Quit
		case tea.KeyCtrlH:
			m.showHelp = !m.showHelp
			return m, nil
		case tea.KeyEnter:
			userInput := m.textarea.Value()
			if strings.TrimSpace(userInput) != "" {
				m.metrics.RecordTurn()
			}
			if strings.TrimSpace(userInput) == "/help" {
				m.showHelp = !m.showHelp
				state := "Help hidden."
				if m.showHelp {
					state = "Help shown."
				}
				m.messages = append(m.messages, message{sender: systemSender, content: state})
				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
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
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
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
								m.refreshPreviewPane()
							}
						}
					} else {
						m.messages = append(m.messages, message{sender: systemSender, content: "No assistant message to save."})
					}
				} else {
					m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /save <filename>"})
				}
				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/edit-values") {
				cmdRun, err := m.startDiffCommand(DiffModeValues, userInput)
				if err != nil {
					return m, nil
				}
				if cmdRun != nil {
					return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmdRun)
				}
				return m, nil
			}
			if strings.HasPrefix(userInput, "/edit-yaml") {
				cmdRun, err := m.startDiffCommand(DiffModeManifest, userInput)
				if err != nil {
					return m, nil
				}
				if cmdRun != nil {
					return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmdRun)
				}
				return m, nil
			}
			if strings.HasPrefix(userInput, "/gen-deploy") {
				cmdRun, err := m.startGenerationCommand(GenerationTypeDeployment, userInput)
				if err != nil {
					return m, nil
				}
				if cmdRun != nil {
					return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmdRun)
				}
				return m, nil
			}
			if strings.HasPrefix(userInput, "/gen-helm") {
				cmdRun, err := m.startGenerationCommand(GenerationTypeHelmChart, userInput)
				if err != nil {
					return m, nil
				}
				if cmdRun != nil {
					return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmdRun)
				}
				return m, nil
			}
			if strings.TrimSpace(userInput) == "/cancel" {
				cancelled := false
				if m.pendingDiff != nil {
					cancelled = m.cancelDiffSession() || cancelled
				}
				if m.pendingGeneration != nil {
					cancelled = m.cancelGenerationSession() || cancelled
				}
				if !cancelled {
					m.messages = append(m.messages, message{sender: systemSender, content: "ℹ️ Nothing to cancel."})
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()
				}
				m.textarea.Reset()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/diag-pod ") {
				parts := strings.Fields(userInput)
				if len(parts) >= 2 {
					ns, _ := GetCurrentNamespace()
					pod := parts[1]

					m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🔍 Running diagnostic plan for pod '%s' in namespace '%s'...", pod, ns)})
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()

					results, err := DiagnosePodNotReady(pod, ns)
					if err != nil {
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error running diagnostics: %v", err)})
						m.textarea.Reset()
						m.chatViewport.SetContent(m.renderMessages())
						m.chatViewport.GotoBottom()
						return m, nil
					}

					// Display diagnostic outputs and heuristic notes
					var diagnosticSummary strings.Builder
					diagnosticSummary.WriteString("🧪 Diagnostic Results:\n\n")

					for _, r := range results {
						m.messages = append(m.messages, message{sender: execSender, content: "$ " + r.Command})
						m.messages = append(m.messages, message{sender: systemSender, content: r.Output})
						for _, note := range r.Notes {
							m.messages = append(m.messages, message{sender: systemSender, content: "💡 " + note})
						}

						// Build summary for LLM analysis
						diagnosticSummary.WriteString(fmt.Sprintf("Command: %s\n", r.Command))
						if r.Output != "" {
							// Truncate output for LLM to avoid token limits
							output := r.Output
							if len(output) > 2000 {
								output = output[:2000] + "\n...(truncated)..."
							}
							diagnosticSummary.WriteString(fmt.Sprintf("Output: %s\n", output))
						}
						if len(r.Notes) > 0 {
							diagnosticSummary.WriteString(fmt.Sprintf("Heuristic Notes: %s\n", strings.Join(r.Notes, "; ")))
						}
						diagnosticSummary.WriteString("\n")
					}

					// Generate LLM synthesis prompt
					analysisPrompt := fmt.Sprintf(`Based on the diagnostic outputs above for pod '%s' in namespace '%s', please provide a concise analysis with:

1. **Root Cause**: What is likely causing the pod issues?
2. **Next Steps**: What specific actions should be taken to resolve this?

Diagnostic Data:
%s

Please be specific and actionable in your recommendations.`, pod, ns, diagnosticSummary.String())

					m.messages = append(m.messages, message{sender: user, content: analysisPrompt})
					history := append([]message(nil), m.messages...)
					m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
					m.chatViewport.SetContent(m.renderMessages())
					m.textarea.Reset()
					m.chatViewport.GotoBottom()
					m.resetLiveTokens()
					cmd = generateStreamCmd(m, history, m.generationModel)
					return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmd)
				} else {
					m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /diag-pod <pod-name>"})
					m.textarea.Reset()
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()
					return m, nil
				}
			}
			if strings.HasPrefix(userInput, "/model") {
				fields := strings.Fields(userInput)
				sub := ""
				if len(fields) > 1 {
					sub = fields[1]
				}

				switch sub {
				case "", "list":
					models, err := ListModels()
					if err != nil {
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error getting models: %v", err)})
					} else {
						current := fmt.Sprintf("Chat: %s\nGeneration: %s", m.ollamaModel, m.generationModel)
						m.messages = append(m.messages, message{sender: assist, content: "Available models:\n" + strings.Join(models, "\n") + "\n\n" + current})
					}
				case "set":
					if len(fields) < 3 {
						m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /model set [chat|generation] <name>"})
						break
					}
					scope := ""
					nameIndex := 2
					if strings.EqualFold(fields[2], "chat") || strings.EqualFold(fields[2], "generation") || strings.EqualFold(fields[2], "gen") || strings.EqualFold(fields[2], "diff") {
						scope = strings.ToLower(fields[2])
						nameIndex = 3
					}
					if nameIndex >= len(fields) {
						m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /model set [chat|generation] <name>"})
						break
					}
					modelName := strings.Join(fields[nameIndex:], " ")
					if err := UpdateModelInConfig(scope, modelName); err != nil {
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Failed to update config: %v", err)})
						break
					}
					if scope == "generation" || scope == "gen" || scope == "diff" {
						m.generationModel = modelName
						if m.config != nil {
							m.config.Models.Generation = modelName
						}
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Generation model set to %s", modelName)})
					} else {
						resolved, status, err := resolveModel(modelName, true)
						if err != nil {
							m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error selecting model: %v", err)})
						} else {
							m.ollamaModel = resolved
							if m.config != nil {
								m.config.Models.Chat = resolved
							}
							if strings.TrimSpace(status) != "" {
								m.messages = append(m.messages, message{sender: systemSender, content: status})
							}
							m.messages = append(m.messages, message{sender: assist, content: fmt.Sprintf("Chat model set to %s", resolved)})
						}
					}
				default:
					// Legacy shorthand: /model <name>
					modelName := strings.TrimSpace(strings.TrimPrefix(userInput, "/model"))
					if modelName == "" {
						m.messages = append(m.messages, message{sender: systemSender, content: "Usage: /model list | /model set [chat|generation] <name>"})
						break
					}
					if err := UpdateModelInConfig("chat", modelName); err != nil {
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Failed to update config: %v", err)})
						break
					}
					resolved, status, err := resolveModel(modelName, true)
					if err != nil {
						m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error selecting model: %v", err)})
					} else {
						m.ollamaModel = resolved
						if m.config != nil {
							m.config.Models.Chat = resolved
						}
						if strings.TrimSpace(status) != "" {
							m.messages = append(m.messages, message{sender: systemSender, content: status})
						}
						m.messages = append(m.messages, message{sender: assist, content: fmt.Sprintf("Chat model set to %s", resolved)})
					}
				}

				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
			if userInput == "/ctx" {
				summary, err := BuildContextSummary()
				if err != nil {
					m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Error getting context: %v", err)})
				} else {
					m.messages = append(m.messages, message{sender: assist, content: fmt.Sprintf("🔍 Current Context:\n%s", summary.RenderedOneLiner)})
				}
				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/ns set ") {
				parts := strings.SplitN(userInput, " ", 3)
				if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
					newNs := strings.TrimSpace(parts[2])
					// Switch namespace using kubectl
					setNsCmd := fmt.Sprintf("kubectl config set-context --current --namespace=%s", newNs)
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + setNsCmd})
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()
					m.namespace = newNs
					m.beginCommandExecution(setNsCmd)
					cmd = execCmd(setNsCmd, m.program)
					m.textarea.Reset()
					return m, cmd
				} else {
					m.messages = append(m.messages, message{sender: assist, content: "Invalid /ns command. Usage: /ns set <namespace>"})
					m.textarea.Reset()
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()
					return m, nil
				}
			}
			if strings.HasPrefix(userInput, "/metrics") {
				table := m.metrics.Table()
				m.messages = append(m.messages, message{sender: systemSender, content: table})
				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
			if strings.HasPrefix(userInput, "/resolve") {
				note := strings.TrimSpace(strings.TrimPrefix(userInput, "/resolve"))
				m.metrics.RecordResolution()
				msg := "✅ Marked task as resolved."
				if note != "" {
					msg = fmt.Sprintf("%s Note: %s", msg, note)
				}
				m.messages = append(m.messages, message{sender: systemSender, content: msg})
				m.textarea.Reset()
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
				return m, nil
			}
			trimmed := strings.TrimSpace(userInput)
			if trimmed == "" {
				return m, nil
			}

			// Track user input for smart refresh
			m.trackUserInput()

			// Update intelligence with user input
			if m.intelligentUI != nil && m.currentContext != nil {
				err := m.intelligentUI.UpdateIntelligence(trimmed, m.currentContext)
				if err == nil {
					m.riskLevel = m.intelligentUI.GetCurrentRiskLevel()
					m.lastIntelligenceUpdate = time.Now()

					// Show intelligent suggestions if confidence is low
					if session := m.intelligentUI.GetCurrentSession(); session != nil && session.Confidence < 0.7 {
						suggestions := m.intelligentUI.FormatSuggestions()
						if suggestions != "" {
							m.messages = append(m.messages, message{sender: systemSender, content: "🤖 AI Suggestions available - press F1-F3 to use them"})
						}
					}

					// Get predictive suggestions if available
					if PredictiveIntelligence != nil {
						predictions := PredictiveIntelligence.PredictNextActions(m.currentContext, trimmed)
						if len(predictions) > 0 {
							m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🔮 %d predictive suggestions available", len(predictions))})
						}
					}
				}
			}

			m.command = ""
			m.currentPlan = nil
			m.refreshPreviewPane()
			m.messages = append(m.messages, message{sender: user, content: trimmed})
			history := append([]message(nil), m.messages...)
			m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
			m.chatViewport.SetContent(m.renderMessages())
			m.textarea.Reset()
			m.chatViewport.GotoBottom()
			m.resetLiveTokens()
			cmd = generateStreamCmd(m, history, m.ollamaModel)
		case tea.KeyCtrlE:
			if m.pendingDiff != nil && m.pendingDiff.Phase() == DiffPhasePreview {
				if err := m.applyPendingDiff(); err != nil {
					m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Failed to apply diff: %v", err)})
					m.chatViewport.SetContent(m.renderMessages())
					m.chatViewport.GotoBottom()
				}
				m.textarea.Reset()
				return m, nil
			}

			if m.awaitingTypedConfirm != nil {
				// For dangerous commands requiring "yes" confirmation
				userInput := strings.TrimSpace(m.textarea.Value())
				if strings.ToLower(userInput) == "yes" {
					realCmd := m.awaitingTypedConfirm.Original
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + realCmd})
					m.chatViewport.SetContent(m.renderMessages())
					m.metrics.RecordConfirmation()
					m.beginCommandExecution(realCmd)
					cmd = execCmd(realCmd, m.program)
					m.awaitingTypedConfirm = nil
					m.textarea.Reset()
				} else {
					m.messages = append(m.messages, message{sender: systemSender, content: "⚠️ Dangerous command cancelled. Type 'yes' and press Ctrl+E to confirm."})
					m.chatViewport.SetContent(m.renderMessages())
				}
				break
			}

			if m.awaitingSecondConfirm != nil {
				// Execute original for real
				realCmd := m.awaitingSecondConfirm.Original
				m.messages = append(m.messages, message{sender: execSender, content: "$ " + realCmd})
				m.chatViewport.SetContent(m.renderMessages())
				m.metrics.RecordConfirmation()
				m.beginCommandExecution(realCmd)
				cmd = execCmd(realCmd, m.program)
				m.awaitingSecondConfirm = nil
				break
			}

			if m.command != "" {
				plan := BuildPreExecPlan(m.command)
				m.currentPlan = &plan
				m.refreshPreviewPane()
				if plan.RequireTypedConfirm {
					m.metrics.RecordSafetyBlock()
				}
				m.metrics.RecordConfirmation()

				// Show comprehensive safety report
				m.messages = append(m.messages, message{sender: systemSender, content: plan.GetSafetyReport()})
				m.messages = append(m.messages, message{sender: systemSender, content: plan.HumanPreview()})
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()

				var cmds []tea.Cmd

				// Run preview checks first
				if len(plan.Checks) > 0 {
					m.messages = append(m.messages, message{sender: systemSender, content: "🧪 Running validation checks..."})
					m.chatViewport.SetContent(m.renderMessages())

					for _, check := range plan.Checks {
						m.messages = append(m.messages, message{sender: execSender, content: "$ " + check.Cmd})
						m.chatViewport.SetContent(m.renderMessages())
						cmds = append(cmds, runPreviewCheck(check, m.program))
					}
				}

				// Run the first command (usually dry-run)
				if plan.FirstRunCommand != "" {
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + plan.FirstRunCommand})
					m.chatViewport.SetContent(m.renderMessages())
					m.beginCommandExecution(plan.FirstRunCommand)
					cmds = append(cmds, execCmd(plan.FirstRunCommand, m.program))
				}

				// Set up confirmation workflow
				if plan.RequireTypedConfirm {
					m.messages = append(m.messages, message{sender: systemSender, content: "🚨 DANGEROUS COMMAND! Type 'yes' and press Ctrl+E to confirm execution."})
					m.awaitingTypedConfirm = &plan
				} else if plan.RequireSecondConfirm {
					m.messages = append(m.messages, message{sender: systemSender, content: "🔄 Press Ctrl+E again to APPLY for real, or edit the command first."})
					m.awaitingSecondConfirm = &plan
				}

				if len(cmds) > 0 {
					cmd = tea.Sequence(cmds...)
				}
			}
		case tea.KeyCtrlK:
			m.command = ""
			m.currentPlan = nil
			m.refreshPreviewPane()
		}

	case ollamaStreamMsg:
		last := len(m.messages) - 1
		if m.messages[last].content == waitingMessage {
			m.messages[last].content = ""
		}
		chunk := string(msg)
		m.messages[last].content += chunk
		if m.pendingDiff != nil && m.pendingDiff.Phase() == DiffPhaseAwaiting {
			m.pendingDiff.AppendResponse(chunk)
		}
		if m.pendingGeneration != nil && m.pendingGeneration.Phase() == GenerationPhaseAwaiting {
			m.pendingGeneration.AppendResponse(chunk)
		}
		m.liveTokens += estimateTokens(chunk)
		if m.liveTokens > maxLiveTokens {
			m.liveTokens = maxLiveTokens
		}
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()

	case ollamaStreamDoneMsg:
		last := len(m.messages) - 1
		m.liveTokens = 0

		if m.pendingDiff != nil && m.pendingDiff.Phase() == DiffPhaseAwaiting {
			m.handleDiffCompletion()
			return m, nil
		}

		if m.pendingGeneration != nil && m.pendingGeneration.Phase() == GenerationPhaseAwaiting {
			m.handleGenerationCompletion()
			return m, nil
		}

		assistantReply := m.messages[last].content

		if m.agentMode && m.agentState == "thinking" {
			if action := parseAction(assistantReply); action != "" {
				if isActionWhitelisted(action) {
					m.agentState = "acting"
					m.messages = append(m.messages, message{sender: execSender, content: "$ " + action})
					m.beginCommandExecution(action)
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
				m.currentPlan = nil
				m.refreshPreviewPane()
				break
			}
			m.command = parseCommandFromResponse(assistantReply)
			m.refreshPreviewPane()
			if m.command != "" {
				m.metrics.RecordSuggestion()
			}
			if m.command == "" {
				if strings.TrimSpace(assistantReply) == "" {
					m.messages[last].content = "I didn’t receive any text from the model."
				}
				m.messages = append(m.messages, message{sender: assist, content: "I could not find a runnable command in that response."})
				m.chatViewport.SetContent(m.renderMessages())
				m.chatViewport.GotoBottom()
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
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.refreshOutputPane()

	case stderrMsg:
		m.stderrContent[msg.cmd] += msg.out + "\n"
		m.messages = append(m.messages, message{sender: systemSender, content: "stderr: " + msg.out})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.refreshOutputPane()

	case execDoneMsg:
		// Learn from command execution for predictive intelligence
		if PredictiveIntelligence != nil && m.currentContext != nil {
			userInput := ""
			if len(m.messages) > 1 {
				// Find the last user input
				for i := len(m.messages) - 1; i >= 0; i-- {
					if m.messages[i].sender == user {
						userInput = m.messages[i].content
						break
					}
				}
			}
			
			success := msg.err == nil
			duration := time.Since(time.Now()) // This would be actual execution time in real implementation
			PredictiveIntelligence.LearnFromInteraction(userInput, m.currentContext, msg.cmd, success, duration)
		}

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
			m.chatViewport.SetContent(m.renderMessages())
			m.textarea.Reset()
			m.chatViewport.GotoBottom()
			cmd = generateStreamCmd(m, history, m.ollamaModel)
			m.agentState = "thinking"

		} else if msg.err != nil {
			// Command failed
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("Command '%s' failed: %v", msg.cmd, msg.err)})

			if stderr, ok := m.stderrContent[msg.cmd]; ok && stderr != "" {
				// There was stderr, trigger self-correction
				correctionPrompt := fmt.Sprintf("The command `%s` failed with the error:\n```\n%s\n```\nCan you fix it?", msg.cmd, stderr)
				m.metrics.RecordCorrection()

				m.messages = append(m.messages, message{sender: user, content: correctionPrompt})
				history := append([]message(nil), m.messages...)
				m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
				m.chatViewport.SetContent(m.renderMessages())
				m.textarea.Reset()
				m.chatViewport.GotoBottom()
				m.resetLiveTokens()
				cmd = generateStreamCmd(m, history, m.generationModel)
			}
		}
		m.currentPlan = nil
		m.command = ""
		m.refreshPreviewPane()
		m.refreshOutputPane()

	case previewCheckDoneMsg:
		// Handle preview check completion
		m.previewCheckResults[msg.check.Name] = msg

		if msg.err != nil {
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: fmt.Sprintf("❌ %s failed: %v", msg.check.Name, msg.err),
			})
		} else {
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: fmt.Sprintf("✅ %s passed", msg.check.Name),
			})
		}
		m.metrics.RecordValidation(msg.err == nil)

		// Show limited output for debugging
		if msg.out != "" && len(msg.out) > 0 {
			output := msg.out
			if len(output) > 500 {
				output = output[:500] + "...(truncated)"
			}
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: fmt.Sprintf("Output:\n```\n%s\n```", output),
			})
		}

	case validationFailedMsg:
		// Handle command validation failure - trigger self-correction
		m.messages = append(m.messages, message{
			sender:  systemSender,
			content: fmt.Sprintf("❌ Command validation failed: %v", msg.err),
		})
		m.metrics.RecordValidation(false)

		// Generate correction prompt
		correctionPrompt := generateCorrectionPrompt(msg.cmd, msg.stderr)
		m.metrics.RecordCorrection()

		m.messages = append(m.messages, message{sender: user, content: correctionPrompt})
		history := append([]message(nil), m.messages...)
		m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
		m.chatViewport.SetContent(m.renderMessages())
		m.textarea.Reset()
		m.chatViewport.GotoBottom()
		cmd = generateStreamCmd(m, history, m.generationModel)
	}

	return m, tea.Batch(tacmd, chatcmd, prevcmd, outcmd, cmd, contextCmd, tickCmd)
}

func (m *model) View() string {
	// Apply appropriate chroming for preview when showing diffs
	if m.rightTopMode == rightPaneDiff {
		m.previewViewport.Style = m.styles.diffStyle
	} else {
		m.previewViewport.Style = m.styles.viewportStyle
	}

	header := m.styles.headerStyle.Render(" KubeMage ")
	chatSection := m.paneView("Chat", m.chatViewport.View())
	previewSection := m.paneView(m.previewTitle(), m.previewViewport.View())
	outputSection := m.paneView("Output / Logs", m.outputViewport.View())

	var body string
	switch m.layout {
	case layoutThreePane:
		right := lipgloss.JoinVertical(lipgloss.Left, previewSection, outputSection)
		body = lipgloss.JoinHorizontal(lipgloss.Top, chatSection, right)
	case layoutVerticalSplit:
		right := lipgloss.JoinVertical(lipgloss.Left, previewSection, outputSection)
		body = lipgloss.JoinHorizontal(lipgloss.Top, chatSection, right)
	case layoutHorizontalSplit:
		bottom := lipgloss.JoinHorizontal(lipgloss.Top, previewSection, outputSection)
		body = lipgloss.JoinVertical(lipgloss.Left, chatSection, bottom)
	case layoutChatOnly:
		body = chatSection
	}

	// Refresh intelligence if needed
	if m.shouldRefreshIntelligence() {
		m.refreshIntelligenceAsync()
	}

	hints := m.renderCommandHints()
	intelligence := m.renderIntelligencePanel()
	input := m.styles.inputWrapper.Render(m.textarea.View())
	help := m.renderHelpBlock()
	context := m.renderEnhancedContextFooter()

	sections := []string{header, body}
	if hints != "" {
		sections = append(sections, hints)
	}
	if intelligence != "" {
		sections = append(sections, intelligence)
	}
	sections = append(sections, input)
	if help != "" {
		sections = append(sections, help)
	}
	if context != "" {
		sections = append(sections, context)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) paneView(title, body string) string {
	if strings.TrimSpace(body) == "" {
		body = "──"
	}
	titleBlock := m.styles.hintKeyStyle.Render(title)
	return lipgloss.JoinVertical(lipgloss.Left, titleBlock, body)
}

func (m *model) previewTitle() string {
	if m.rightTopMode == rightPaneDiff {
		return "Diff Preview"
	}
	if m.currentPlan != nil {
		return "Plan Preview"
	}
	if strings.TrimSpace(m.command) != "" {
		return "Command Preview"
	}
	return "Preview"
}

func (m *model) renderContextFooter() string {
	ctx := strings.TrimSpace(m.ctxName)
	if ctx == "" {
		ctx = "(ctx)"
	}
	ns := strings.TrimSpace(m.namespace)
	if ns == "" {
		ns = "(ns)"
	}
	usr := strings.TrimSpace(m.rbacUser)
	if usr == "" {
		usr = "(user)"
	}
	modelLabel := modelFooterLabel(m.ollamaModel)
	tokens := m.liveTokens
	if tokens > maxLiveTokens {
		tokens = maxLiveTokens
	}
	tokenLabel := fmt.Sprintf("tokens:%d", tokens)
	origin := time.Now()
	if !m.lastFooterUpdate.IsZero() {
		origin = m.lastFooterUpdate
	}
	timeLabel := origin.Format("15:04:05")
	parts := []string{
		fmt.Sprintf("ctx:%s", ctx),
		fmt.Sprintf("ns:%s", ns),
		fmt.Sprintf("user:%s", usr),
		fmt.Sprintf("model:%s", modelLabel),
		tokenLabel,
		fmt.Sprintf("time:%s", timeLabel),
	}
	line := strings.Join(parts, "  ")
	style := m.styles.contextStyle
	if strings.Contains(strings.ToLower(ns), "prod") {
		style = m.styles.contextAlert
	}
	return style.Render(line)
}

func (m *model) refreshPreviewPane() {
	var sections []string
	mode := rightPaneText

	// Use adaptive UI for layout optimization if available
	if AdaptiveUI != nil && m.currentContext != nil {
		content := ""
		if len(sections) > 0 {
			content = strings.Join(sections, "\n")
		}
		adaptedLayout := AdaptiveUI.AdaptLayout(content, m.currentContext, m.layout)
		if adaptedLayout != m.layout {
			m.layout = adaptedLayout
			m.updateLayout()
		}
	}

	if m.pendingDiff != nil {
		switch m.pendingDiff.Phase() {
		case DiffPhaseAwaiting:
			sections = append(sections, "Generating diff preview…")
		case DiffPhasePreview:
			if m.pendingDiff.ParsedDiff != nil {
				mode = rightPaneDiff
				sections = append(sections, m.pendingDiff.ParsedDiff.RenderColoredDiff())
				if adds, removes := m.pendingDiff.ParsedDiff.GetDiffStats(); adds+removes > 0 {
					sections = append(sections, fmt.Sprintf("Summary: +%d / -%d", adds, removes))
				}
			}
		case DiffPhaseApplied:
			sections = append(sections, "Diff applied. Ready for validation.")
		}
	}

	if len(sections) == 0 && m.currentPlan != nil {
		sections = append(sections, m.currentPlan.GetSafetyReport())
		sections = append(sections, m.currentPlan.HumanPreview())
	}

	if len(sections) == 0 && strings.TrimSpace(m.command) != "" {
		// Enhanced command preview with highlighting and explanation
		highlightedCmd := m.highlightCommand(m.command)
		sections = append(sections, fmt.Sprintf("Pending command:\n$ %s", highlightedCmd))

		// Add command explanation
		explanation := m.explainCommand(m.command)
		if explanation != "" {
			sections = append(sections, explanation)
		}
	}

	if len(sections) == 0 {
		sections = append(sections, "No preview ready. Ask for a command or start a diff edit.")
	}

	m.rightTopMode = mode
	m.previewViewport.SetContent(strings.Join(sections, "\n\n"))
}

func (m *model) refreshOutputPane() {
	var sb strings.Builder
	if m.activeCommand != "" {
		sb.WriteString(fmt.Sprintf("$ %s\n", m.activeCommand))
		stdout := strings.TrimSpace(m.stdoutContent[m.activeCommand])
		stderr := strings.TrimSpace(m.stderrContent[m.activeCommand])
		if stdout != "" {
			sb.WriteString("── stdout ──\n")
			if len(stdout) > m.config.Truncation.Message {
				stdout = stdout[:m.config.Truncation.Message] + "\n(...truncated...)"
			}
			sb.WriteString(stdout)
			sb.WriteString("\n")
		}
		if stderr != "" {
			sb.WriteString("── stderr ──\n")
			if len(stderr) > m.config.Truncation.Message {
				stderr = stderr[:m.config.Truncation.Message] + "\n(...truncated...)"
			}
			sb.WriteString(stderr)
			sb.WriteString("\n")
		}
		if stdout == "" && stderr == "" {
			sb.WriteString("(no output yet)")
		}
	} else {
		sb.WriteString("No command running. Press Ctrl+E for a dry-run.")
	}

	m.outputViewport.SetContent(sb.String())
}

func (m *model) beginCommandExecution(command string) {
	for k := range m.stdoutContent {
		delete(m.stdoutContent, k)
	}
	for k := range m.stderrContent {
		delete(m.stderrContent, k)
	}
	m.activeCommand = command
	m.stdoutContent[command] = ""
	m.stderrContent[command] = ""
	m.refreshOutputPane()
}

func (m *model) resetLiveTokens() {
	m.liveTokens = 0
}

func requestContextSummary() tea.Cmd {
	return func() tea.Msg {
		summary, err := BuildContextSummary()
		return contextSummaryMsg{summary: summary, err: err}
	}
}

func scheduleContextRefresh() tea.Cmd {
	return tea.Tick(contextRefreshInterval, func(time.Time) tea.Msg {
		summary, err := BuildContextSummary()
		return contextSummaryMsg{summary: summary, err: err}
	})
}

func scheduleClockTick() tea.Cmd {
	return tea.Tick(clockTickInterval, func(t time.Time) tea.Msg {
		return clockTickMsg(t)
	})
}

func estimateTokens(chunk string) int {
	count := len(strings.Fields(chunk))
	if count == 0 {
		count = len([]rune(chunk)) / 4
	}
	if count < 0 {
		count = 0
	}
	return count
}

func modelFooterLabel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "(model)"
	}
	parts := strings.Split(trimmed, ":")
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return fmt.Sprintf("%s@%s", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	return trimmed
}

func (m *model) renderMessages() string {
	var history string
	for _, msg := range m.messages {
		var senderStyle lipgloss.Style
		senderText := msg.sender
		content := RedactText(msg.content)
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
		if len(content) > m.config.Truncation.Message {
			content = content[:m.config.Truncation.Message] + "\n(...truncated...)"
		}
		history += senderStyle.Render(senderText) + ": " + content + "\n"
	}
	return history
}

func (m *model) renderStatusLine() string {
	var snap MetricsSnapshot
	if m.metrics != nil {
		snap = m.metrics.Snapshot()
	}
	status := fmt.Sprintf("Chat: %s • Gen: %s", m.ollamaModel, m.generationModel)
	if m.metrics != nil {
		status = fmt.Sprintf("%s • TSR %.0f%% • CAR %.0f%% • EAR %.0f%%", status, snap.TSR*100, snap.CAR*100, snap.EAR*100)
	}
	status = fmt.Sprintf("%s • Ctrl+H help", status)
	return m.styles.statusStyle.Render(status)
}

func (m *model) renderCommandHints() string {
	input := strings.TrimSpace(m.textarea.Value())
	lowerInput := strings.ToLower(input)
	var hints []commandHint
	if strings.HasPrefix(input, "/") {
		for _, hint := range commandPalette {
			if strings.HasPrefix(strings.ToLower(hint.Trigger), lowerInput) {
				hints = append(hints, hint)
			}
		}
	} else {
		maxHints := 4
		if len(commandPalette) < maxHints {
			maxHints = len(commandPalette)
		}
		hints = append(hints, commandPalette[:maxHints]...)
	}
	if len(hints) > 6 {
		hints = hints[:6]
	}

	if len(hints) == 0 {
		return ""
	}

	lines := make([]string, 0, len(hints)+1)
	lines = append(lines, m.styles.hintDescStyle.Render("Commands"))
	for _, hint := range hints {
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			m.styles.hintKeyStyle.Render(hint.Trigger),
			" ",
			m.styles.hintDescStyle.Render(hint.Description),
		)
		lines = append(lines, line)
	}

	return m.styles.hintBoxStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) renderHelpBlock() string {
	if !m.showHelp {
		return ""
	}

	helpSections := []string{
		m.styles.hintKeyStyle.Render("🎮 Enhanced Controls:"),
		"Enter: Ask question • Ctrl+E: Execute command • Ctrl+K: Clear preview",
		"Ctrl+H: Toggle help • F2: Change layout • Ctrl+C: Exit",
		"",
		m.styles.hintKeyStyle.Render("🤖 AI Intelligence Hotkeys:"),
		"F1-F3: Execute AI suggestions • F4: Toggle suggestions • F5: Toggle insights",
		"F12: Force refresh intelligence",
		"",
		m.styles.hintKeyStyle.Render("💬 Slash Commands:"),
		"/model set chat <name> • /edit-yaml <file> <instruction> • /metrics",
		"/resolve [note] • /agent • /diag-pod <name> • /ctx • /ns set <namespace>",
		"",
		m.styles.hintKeyStyle.Render("🎨 Features:"),
		"• Real-time cluster health monitoring with risk indicators",
		"• Intelligent command suggestions based on context",
		"• Syntax highlighting and command explanations",
		"• Multi-step diagnostic workflows with pattern recognition",
		"• Responsive layout adapting to screen size and content",
		"",
		m.styles.warningIndicator.Render("⚠️ Prerequisites:"),
		"Ensure 'ollama serve' is running locally or set OLLAMA_HOST",
		"Active kubectl context required for cluster operations",
	}

	// Style the help content
	helpContent := strings.Join(helpSections, "\n")
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4C566A")).
		Padding(1).
		Margin(1)

	return helpStyle.Render(helpContent)
}

func (m *model) renderCommandPreview() string {
	if strings.TrimSpace(m.command) == "" {
		return ""
	}
	title := m.styles.hintKeyStyle.Render("Pending command")
	body := m.styles.hintDescStyle.Render(m.command)
	return m.styles.hintBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))
}

func (m *model) highlightYAML(content string) string {
	keywords := []string{"apiVersion:", "kind:", "metadata:", "spec:", "labels:", "selector:", "template:", "containers:", "image:", "ports:", "name:", "replicas:", "app:"}
	highlightedContent := content
	for _, keyword := range keywords {
		highlightedContent = strings.ReplaceAll(highlightedContent, keyword, m.styles.yamlKeyStyle.Render(keyword))
	}
	return highlightedContent
}

// highlightCommand provides syntax highlighting for kubectl/helm commands
func (m *model) highlightCommand(command string) string {
	if strings.TrimSpace(command) == "" {
		return command
	}

	// Split command into parts for highlighting
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return command
	}

	var highlighted []string

	for i, part := range parts {
		switch {
		case i == 0: // Command name
			if strings.HasPrefix(part, "kubectl") {
				highlighted = append(highlighted, m.styles.execStyle.Render(part))
			} else if strings.HasPrefix(part, "helm") {
				highlighted = append(highlighted, m.styles.assistStyle.Render(part))
			} else {
				highlighted = append(highlighted, m.styles.systemStyle.Render(part))
			}

		case strings.HasPrefix(part, "--"): // Flags
			highlighted = append(highlighted, m.styles.hintKeyStyle.Render(part))

		case strings.HasPrefix(part, "-"): // Short flags
			highlighted = append(highlighted, m.styles.hintKeyStyle.Render(part))

		case part == "get" || part == "describe" || part == "logs" || part == "delete" || part == "apply":
			// Kubectl verbs
			highlighted = append(highlighted, m.styles.yamlKeyStyle.Render(part))

		case part == "install" || part == "upgrade" || part == "uninstall" || part == "lint":
			// Helm verbs
			highlighted = append(highlighted, m.styles.yamlKeyStyle.Render(part))

		case strings.Contains(part, "="): // Key-value pairs
			kvParts := strings.SplitN(part, "=", 2)
			if len(kvParts) == 2 {
				highlighted = append(highlighted,
					m.styles.hintKeyStyle.Render(kvParts[0]) + "=" + m.styles.hintDescStyle.Render(kvParts[1]))
			} else {
				highlighted = append(highlighted, part)
			}

		default:
			// Regular arguments
			highlighted = append(highlighted, m.styles.hintDescStyle.Render(part))
		}
	}

	return strings.Join(highlighted, " ")
}

// explainCommand provides intelligent explanations for commands
func (m *model) explainCommand(command string) string {
	if strings.TrimSpace(command) == "" {
		return ""
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	var explanations []string

	// Command-specific explanations
	if strings.HasPrefix(command, "kubectl") {
		if len(parts) > 1 {
			switch parts[1] {
			case "get":
				explanations = append(explanations, "📄 Retrieves and displays resources")
				if strings.Contains(command, "--dry-run") {
					explanations = append(explanations, "🔍 Dry-run mode - no changes will be made")
				}
			case "describe":
				explanations = append(explanations, "🔍 Shows detailed information about resources")
			case "logs":
				explanations = append(explanations, "📋 Displays container logs")
				if strings.Contains(command, "--previous") {
					explanations = append(explanations, "⏮️ Shows logs from previous container instance")
				}
			case "apply":
				explanations = append(explanations, "🚀 Creates or updates resources from configuration")
				if strings.Contains(command, "--dry-run") {
					explanations = append(explanations, "✅ Safe dry-run - validates without applying")
				}
			case "delete":
				explanations = append(explanations, "🗑️ Removes resources from cluster")
				explanations = append(explanations, "⚠️ DESTRUCTIVE OPERATION")
			case "top":
				explanations = append(explanations, "📊 Shows resource usage metrics")
			}
		}
	} else if strings.HasPrefix(command, "helm") {
		if len(parts) > 1 {
			switch parts[1] {
			case "install":
				explanations = append(explanations, "📦 Installs a new Helm chart")
				if strings.Contains(command, "--dry-run") {
					explanations = append(explanations, "🔍 Dry-run mode - simulates installation")
				}
			case "upgrade":
				explanations = append(explanations, "⬆️ Upgrades an existing Helm release")
			case "uninstall":
				explanations = append(explanations, "🗑️ Removes a Helm release")
				explanations = append(explanations, "⚠️ DESTRUCTIVE OPERATION")
			case "lint":
				explanations = append(explanations, "🔍 Validates chart syntax and best practices")
			case "template":
				explanations = append(explanations, "📝 Renders chart templates locally")
			}
		}
	}

	// Risk assessment
	riskLevel := m.assessCommandRisk(command)
	switch riskLevel {
	case "low":
		explanations = append(explanations, m.styles.healthyIndicator.Render("🟢 Low risk - safe to execute"))
	case "medium":
		explanations = append(explanations, m.styles.warningIndicator.Render("🟡 Medium risk - review carefully"))
	case "high":
		explanations = append(explanations, m.styles.errorIndicator.Render("🔴 High risk - production impact possible"))
	case "critical":
		explanations = append(explanations, m.styles.riskCritStyle.Render("🚨 CRITICAL RISK - potential data loss"))
	}

	if len(explanations) == 0 {
		return ""
	}

	return strings.Join(explanations, "\n")
}

// assessCommandRisk evaluates the risk level of a command
func (m *model) assessCommandRisk(command string) string {
	cmdLower := strings.ToLower(command)

	// Critical risk patterns
	if strings.Contains(cmdLower, "delete") && (strings.Contains(cmdLower, "--all") || strings.Contains(cmdLower, "*")) {
		return "critical"
	}

	// High risk patterns
	if strings.Contains(cmdLower, "delete") || strings.Contains(cmdLower, "uninstall") {
		return "high"
	}

	// Medium risk patterns
	if strings.Contains(cmdLower, "apply") || strings.Contains(cmdLower, "patch") || strings.Contains(cmdLower, "install") || strings.Contains(cmdLower, "upgrade") {
		if strings.Contains(cmdLower, "--dry-run") {
			return "low"
		}
		return "medium"
	}

	// Low risk (read-only operations)
	if strings.Contains(cmdLower, "get") || strings.Contains(cmdLower, "describe") ||
	   strings.Contains(cmdLower, "logs") || strings.Contains(cmdLower, "top") ||
	   strings.Contains(cmdLower, "lint") || strings.Contains(cmdLower, "template") {
		return "low"
	}

	return "medium"
}

func parseCommandFromResponse(response string) string {
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

func ensureWorkspaceInitialized() error {
	if WorkspaceIdx == nil {
		return InitializeWorkspace()
	}
	return nil
}

func ensureValidationPipeline() {
	if ValidationPipe == nil {
		InitializeValidation()
	}
}

func (m *model) startDiffCommand(mode DiffMode, input string) (tea.Cmd, error) {
	trimmed := strings.TrimSpace(input)
	usage := "Usage: /edit-values <path> <instruction>"
	if mode == DiffModeManifest {
		usage = "Usage: /edit-yaml <path> <instruction>"
	}

	m.messages = append(m.messages, message{sender: user, content: trimmed})

	if m.pendingDiff != nil && m.pendingDiff.Phase() != DiffPhaseApplied && m.pendingDiff.Phase() != DiffPhaseNone {
		m.messages = append(m.messages, message{sender: systemSender, content: "⚠️ Finish the current diff review before starting another edit."})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, fmt.Errorf("diff session already in progress")
	}

	if err := ensureWorkspaceInitialized(); err != nil {
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Workspace initialization failed: %v", err)})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, err
	}

	parts := strings.SplitN(strings.TrimSpace(input), " ", 3)
	if len(parts) < 3 {
		m.messages = append(m.messages, message{sender: systemSender, content: usage})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, fmt.Errorf("invalid edit command")
	}

	pathArg := strings.TrimSpace(parts[1])
	instruction := strings.TrimSpace(parts[2])
	if pathArg == "" || instruction == "" {
		m.messages = append(m.messages, message{sender: systemSender, content: usage})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, fmt.Errorf("path or instruction missing")
	}
	safeInstruction := RedactText(instruction)

	normalizedPath := WorkspaceIdx.NormalizePath(pathArg)
	absPath := WorkspaceIdx.AbsPath(normalizedPath)
	if _, err := os.Stat(absPath); err != nil {
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ File not found: %s", normalizedPath)})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, err
	}

	currentContent, err := GetFileContent(absPath)
	if err != nil {
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Unable to read %s: %v", normalizedPath, err)})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, err
	}
	redaction := RedactSensitive(currentContent)

	m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("✏️ Generating diff for %s", normalizedPath)})

	prompt := BuildDiffEditPrompt(mode, normalizedPath, redaction.Sanitized, safeInstruction)
	m.pendingDiff = NewDiffSession(mode, normalizedPath, safeInstruction, currentContent, redaction)
	m.refreshPreviewPane()

	history := append([]message(nil), m.messages...)
	history = append(history, message{sender: user, content: prompt})

	m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
	m.textarea.Reset()

	m.resetLiveTokens()
	return generateStreamCmd(m, history, m.generationModel), nil
}

func (m *model) handleDiffCompletion() {
	session := m.pendingDiff
	if session == nil || session.Phase() != DiffPhaseAwaiting {
		return
	}

	raw := session.RawResponse()
	if err := ValidateDiff(raw); err != nil {
		m.messages[len(m.messages)-1] = message{sender: systemSender, content: fmt.Sprintf("⚠️ Diff generation failed: %v", err)}
		m.pendingDiff = nil
		m.refreshPreviewPane()
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		return
	}

	diff, err := ParseUnifiedDiff(raw)
	if err != nil {
		m.messages[len(m.messages)-1] = message{sender: systemSender, content: fmt.Sprintf("⚠️ Unable to parse diff: %v", err)}
		m.pendingDiff = nil
		m.refreshPreviewPane()
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		return
	}

	modified, err := ApplyUnifiedDiff(session.RedactedContent, raw)
	if err != nil {
		m.messages[len(m.messages)-1] = message{sender: systemSender, content: fmt.Sprintf("⚠️ Failed to apply diff preview: %v", err)}
		m.pendingDiff = nil
		m.refreshPreviewPane()
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		return
	}

	session.ParsedDiff = diff
	session.modifiedRedacted = modified
	session.ModifiedContent = RestoreSecrets(modified, session.Redactions)
	m.metrics.RecordEditSuggestion()
	session.SetPhase(DiffPhasePreview)
	m.refreshPreviewPane()

	adds, removes := diff.GetDiffStats()
	preview := diff.RenderColoredDiff()
	m.messages[len(m.messages)-1] = message{sender: systemSender, content: preview}
	m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🧮 Diff summary for %s: +%d / -%d", session.FilePath, adds, removes)})
	m.messages = append(m.messages, message{sender: systemSender, content: "Press Ctrl+E to apply this patch, or type /cancel to discard."})
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
}

func (m *model) applyPendingDiff() error {
	session := m.pendingDiff
	if session == nil || session.Phase() != DiffPhasePreview {
		return fmt.Errorf("no diff ready to apply")
	}

	absPath := WorkspaceIdx.AbsPath(session.FilePath)
	backupPath, err := CreateBackup(absPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if err := WriteFileContent(absPath, session.ModifiedContent); err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}
	m.metrics.RecordEditApplied()

	session.SetPhase(DiffPhaseApplied)
	m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("💾 Applied diff to %s (backup at %s)", session.FilePath, filepath.Base(backupPath))})

	if err := RefreshWorkspace(); err != nil {
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Workspace refresh failed: %v", err)})
	}

	ensureValidationPipeline()
	results := ValidationPipe.ValidateFile(session.FilePath)
	for _, res := range results {
		m.metrics.RecordValidation(res.Success)
	}
	m.messages = append(m.messages, message{sender: systemSender, content: RenderValidationResults(results)})

	if session.Mode == DiffModeManifest {
		m.command = fmt.Sprintf("kubectl apply -f %s", session.FilePath)
		m.refreshPreviewPane()
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("✅ Ready to apply: %s", m.command)})
		m.metrics.RecordSuggestion()
	} else {
		if ok, chartDir := WorkspaceIdx.IsUnderHelmChart(session.FilePath); ok {
			suggestion := fmt.Sprintf("helm upgrade --install %s %s --dry-run --debug", filepath.Base(chartDir), chartDir)
			m.command = suggestion
			m.refreshPreviewPane()
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("✅ Helm dry-run command staged: %s", suggestion)})
			m.metrics.RecordSuggestion()
			if HelmDiffAvailable() {
				valuesArg := session.FilePath
				m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🔍 Helm diff available. Example: helm diff upgrade <release> %s --values %s", chartDir, valuesArg)})
			}
		}
	}

	m.pendingDiff = nil
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
	m.refreshPreviewPane()
	return nil
}

func (m *model) cancelDiffSession() bool {
	if m.pendingDiff == nil {
		return false
	}
	m.pendingDiff = nil
	m.messages = append(m.messages, message{sender: systemSender, content: "ℹ️ Diff editing session cancelled."})
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
	m.refreshPreviewPane()
	return true
}

func (m *model) startGenerationCommand(genType GenerationType, input string) (tea.Cmd, error) {
	trimmed := strings.TrimSpace(input)
	m.messages = append(m.messages, message{sender: user, content: trimmed})

	if m.pendingGeneration != nil && m.pendingGeneration.Phase() != GenerationPhaseCompleted && m.pendingGeneration.Phase() != GenerationPhaseNone {
		m.messages = append(m.messages, message{sender: systemSender, content: "⚠️ A generation workflow is already running."})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, fmt.Errorf("generation in progress")
	}

	if err := ensureWorkspaceInitialized(); err != nil {
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Workspace initialization failed: %v", err)})
		m.chatViewport.SetContent(m.renderMessages())
		m.chatViewport.GotoBottom()
		m.textarea.Reset()
		return nil, err
	}

	var (
		prompt  string
		session *GenerationSession
		summary string
	)

	switch genType {
	case GenerationTypeDeployment:
		opts, err := parseGenDeployCommand(input)
		if err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ %v", err)})
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			m.textarea.Reset()
			return nil, err
		}
		prompt = GetDeploymentGenerationPrompt(opts)
		target := filepath.Join("out", fmt.Sprintf("%s-deploy.yaml", opts.Name))
		session = NewGenerationSession(GenerationTypeDeployment, opts.Name, target, "")
		session.DeployOptions = &opts
		summary = fmt.Sprintf("🧬 Generating deployment for %s (image: %s)", opts.Name, opts.Image)
	case GenerationTypeHelmChart:
		opts, err := parseGenHelmCommand(input)
		if err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ %v", err)})
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			m.textarea.Reset()
			return nil, err
		}
		prompt = GetHelmChartGenerationPrompt(opts)
		targetDir := filepath.Join("out", opts.Name)
		session = NewGenerationSession(GenerationTypeHelmChart, opts.Name, "", targetDir)
		session.HelmOptions = &opts
		summary = fmt.Sprintf("🧬 Generating helm chart %s", opts.Name)
	default:
		return nil, fmt.Errorf("unsupported generation type")
	}

	m.messages = append(m.messages, message{sender: systemSender, content: summary})

	m.pendingGeneration = session
	history := append([]message(nil), m.messages...)
	history = append(history, message{sender: user, content: prompt})

	m.messages = append(m.messages, message{sender: assist, content: waitingMessage})
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
	m.textarea.Reset()

	m.resetLiveTokens()
	return generateStreamCmd(m, history, m.generationModel), nil
}

func (m *model) handleGenerationCompletion() {
	session := m.pendingGeneration
	if session == nil || session.Phase() != GenerationPhaseAwaiting {
		return
	}

	session.SetPhase(GenerationPhasePreview)
	raw := session.RawResponse()

	switch session.Type {
	case GenerationTypeDeployment:
		content := ParseGeneratedContent(raw)
		if strings.TrimSpace(content) == "" {
			m.messages[len(m.messages)-1] = message{sender: systemSender, content: "⚠️ Manifest generation returned empty content."}
			m.pendingGeneration = nil
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			return
		}

		if _, err := os.Stat(session.TargetPath); err == nil {
			backup := fmt.Sprintf("%s.backup.%d", session.TargetPath, time.Now().Unix())
			if err := os.Rename(session.TargetPath, backup); err == nil {
				m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("ℹ️ Previous manifest moved to %s", backup)})
			}
		}

		if err := WriteFileContent(session.TargetPath, content); err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Failed to write manifest: %v", err)})
			m.pendingGeneration = nil
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			return
		}

		if err := RefreshWorkspace(); err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Workspace refresh failed: %v", err)})
		}

		ensureValidationPipeline()
		results := ValidationPipe.ValidateFile(session.TargetPath)
		for _, res := range results {
			m.metrics.RecordValidation(res.Success)
		}
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("💾 Saved manifest to %s", session.TargetPath)})
		m.messages = append(m.messages, message{sender: systemSender, content: RenderValidationResults(results)})
		m.command = fmt.Sprintf("kubectl apply --dry-run=client -f %s", session.TargetPath)
		m.refreshPreviewPane()
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("✅ Dry-run command staged: %s", m.command)})
		m.metrics.RecordSuggestion()

	case GenerationTypeHelmChart:
		files := ParseHelmChartFiles(raw)
		if len(files) == 0 {
			m.messages[len(m.messages)-1] = message{sender: systemSender, content: "⚠️ Helm generation returned no files."}
			m.pendingGeneration = nil
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			return
		}

		if _, err := os.Stat(session.TargetDir); err == nil {
			backup := fmt.Sprintf("%s.backup.%d", session.TargetDir, time.Now().Unix())
			if err := os.Rename(session.TargetDir, backup); err == nil {
				m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("ℹ️ Previous chart moved to %s", backup)})
			}
		}

		chartPath, err := SaveHelmChart(session.Name, files)
		if err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Failed to save chart: %v", err)})
			m.pendingGeneration = nil
			m.chatViewport.SetContent(m.renderMessages())
			m.chatViewport.GotoBottom()
			return
		}

		if err := RefreshWorkspace(); err != nil {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚠️ Workspace refresh failed: %v", err)})
		}

		ensureValidationPipeline()
		results := ValidationPipe.ValidateDirectory(chartPath)
		for _, res := range results {
			m.metrics.RecordValidation(res.Success)
		}
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("💾 Chart files saved under %s", chartPath)})
		m.messages = append(m.messages, message{sender: systemSender, content: RenderValidationResults(results)})
		m.command = fmt.Sprintf("helm upgrade --install %s %s --dry-run --debug", session.Name, chartPath)
		m.refreshPreviewPane()
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("✅ Helm dry-run command staged: %s", m.command)})
		m.metrics.RecordSuggestion()
		if HelmDiffAvailable() {
			m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🔍 Helm diff available. Example: helm diff upgrade %s %s", session.Name, chartPath)})
		}
	}

	session.SetPhase(GenerationPhaseCompleted)
	m.pendingGeneration = nil
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
}

func (m *model) cancelGenerationSession() bool {
	if m.pendingGeneration == nil {
		return false
	}
	m.pendingGeneration = nil
	m.messages = append(m.messages, message{sender: systemSender, content: "ℹ️ Generation workflow cancelled."})
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
	return true
}

func (m *model) buildChatPrompt(history []message) string {
	if len(history) == 0 {
		return ""
	}

	start := 0
	if len(history) > m.config.HistoryLength {
		start = len(history) - m.config.HistoryLength
	}
	trimmed := history[start:]

	var sb strings.Builder
	for _, msg := range trimmed {
		content := strings.TrimSpace(RedactText(msg.content))
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

// Enhanced intelligence methods for TUI

func (m *model) executeAISuggestion(suggestion *AISuggestion) {
	switch suggestion.Type {
	case "diagnostic":
		m.agentMode = true
		m.agentState = "thinking"
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🔍 Executing AI suggestion: %s", suggestion.Title)})
	case "optimization":
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("⚡ Applying optimization: %s", suggestion.Title)})
	case "clarification":
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("❓ Clarifying intent: %s", suggestion.Description)})
	default:
		m.messages = append(m.messages, message{sender: systemSender, content: fmt.Sprintf("🤖 AI Suggestion: %s", suggestion.Action)})
	}

	// Add the suggestion action to textarea for execution
	m.textarea.SetValue(suggestion.Action)
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
}

func (m *model) refreshIntelligence() {
	if m.intelligentUI == nil {
		return
	}

	// Get current context
	context, err := BuildContextSummary()
	if err == nil {
		m.currentContext = context
		m.ctxName = context.Context
		m.namespace = context.Namespace

		// Update cluster health
		if len(context.PodProblemCounts) == 0 {
			m.clusterHealth = "healthy"
		} else if len(context.PodProblemCounts) > 5 {
			m.clusterHealth = "critical"
		} else {
			m.clusterHealth = "warning"
		}
	}

	// Update intelligence with current input using smart caching
	input := strings.TrimSpace(m.textarea.Value())
	if input != "" {
		// Try cache-enabled analysis first
		session, err := Intelligence.AnalyzeIntelligentyWithCache(input, m.currentContext)
		if err == nil && session != nil {
			m.intelligentUI.currentSession = session
			m.riskLevel = m.intelligentUI.calculateUIRiskLevel(session, m.currentContext)
			m.lastIntelligenceUpdate = time.Now()
		} else {
			// Fallback to regular intelligence update
			err := m.intelligentUI.UpdateIntelligence(input, m.currentContext)
			if err == nil {
				m.riskLevel = m.intelligentUI.GetCurrentRiskLevel()
				m.lastIntelligenceUpdate = time.Now()
			}
		}
	}
}

func (m *model) shouldRefreshIntelligence() bool {
	if m.intelligentUI == nil {
		return false
	}

	// Smart activity-based refresh logic
	timeSinceUpdate := time.Since(m.lastIntelligenceUpdate)

	// Force refresh if context changed significantly
	if m.contextChanged() {
		return true
	}

	// Refresh if user is actively typing and enough time has passed
	if m.isUserTyping() && timeSinceUpdate > 5*time.Second {
		return true
	}

	// Refresh if cluster state likely changed (based on recent commands)
	if m.clusterStateChanged() && timeSinceUpdate > 10*time.Second {
		return true
	}

	// Fallback: refresh every 60s for background monitoring (doubled from 30s)
	if timeSinceUpdate > 60*time.Second {
		return true
	}

	return m.intelligentUI.NeedsRefresh()
}

func (m *model) renderIntelligencePanel() string {
	if m.intelligentUI == nil {
		return ""
	}

	var sections []string

	// Quick actions
	if m.showSuggestions {
		if suggestions := m.intelligentUI.FormatSuggestions(); suggestions != "" {
			sections = append(sections, suggestions)
		}
	}

	// Insights
	if m.showInsights {
		if insights := m.intelligentUI.FormatInsights(); insights != "" {
			sections = append(sections, insights)
		}
	}

	// Risk indicator
	if m.intelligentUI.IsHighRisk() {
		riskIndicator := m.styles.riskHighStyle.Render("⚠️ HIGH RISK ENVIRONMENT")
		if m.riskLevel.Level == "critical" {
			riskIndicator = m.styles.riskCritStyle.Render("🚨 CRITICAL RISK")
		}
		sections = append(sections, riskIndicator)
	}

	if len(sections) == 0 {
		return ""
	}

	// Style the panel
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5E81AC")).
		Padding(1).
		Margin(1)

	return panelStyle.Render(strings.Join(sections, "\n\n"))
}

func (m *model) renderEnhancedContextFooter() string {
	if m.intelligentUI == nil {
		return m.renderContextFooter()
	}

	// Build enhanced status with intelligence indicators
	var statusParts []string

	// Context info with health indicator
	if m.currentContext != nil {
		ctxStyle := m.styles.contextStyle
		if strings.Contains(strings.ToLower(m.namespace), "prod") {
			ctxStyle = m.styles.contextAlert
		}

		healthIcon := "🟢"
		healthStyle := m.styles.healthyIndicator
		switch m.clusterHealth {
		case "warning":
			healthIcon = "🟡"
			healthStyle = m.styles.warningIndicator
		case "critical":
			healthIcon = "🔴"
			healthStyle = m.styles.errorIndicator
		}

		statusParts = append(statusParts,
			fmt.Sprintf("ctx:%s", ctxStyle.Render(m.ctxName)),
			fmt.Sprintf("ns:%s", ctxStyle.Render(m.namespace)),
			fmt.Sprintf("health:%s", healthStyle.Render(healthIcon+m.clusterHealth)))
	}

	// Model and intelligence info
	modelStyle := m.styles.statusStyle
	statusParts = append(statusParts, fmt.Sprintf("model:%s", modelStyle.Render(modelFooterLabel(m.ollamaModel))))

	// Risk indicator
	riskIndicator := m.intelligentUI.FormatRiskIndicator()
	statusParts = append(statusParts, fmt.Sprintf("risk:%s", riskIndicator))

	// Intelligence confidence
	if session := m.intelligentUI.GetCurrentSession(); session != nil {
		confStyle := m.styles.confidenceStyle
		confidence := fmt.Sprintf("%.0f%%", session.Confidence*100)
		statusParts = append(statusParts, fmt.Sprintf("ai:%s", confStyle.Render(confidence)))
	}

	// Time
	timeStyle := m.styles.statusStyle
	statusParts = append(statusParts, fmt.Sprintf("time:%s", timeStyle.Render(time.Now().Format("15:04:05"))))

	return m.styles.contextStyle.Render(strings.Join(statusParts, " "))
}

// Smart refresh helper methods

func (m *model) contextChanged() bool {
	if m.currentContext == nil {
		return false
	}

	// Create a hash of the current context
	contextHash := fmt.Sprintf("%s:%s:%d:%d",
		m.currentContext.Context,
		m.currentContext.Namespace,
		len(m.currentContext.PodPhaseCounts),
		len(m.currentContext.PodProblemCounts))

	if m.lastContextHash != contextHash {
		m.lastContextHash = contextHash
		return true
	}

	return false
}

func (m *model) isUserTyping() bool {
	// Check if user input is recent
	timeSinceInput := time.Since(m.lastUserInput)
	return timeSinceInput < m.userTypingTimeout
}

func (m *model) clusterStateChanged() bool {
	// Check if recent commands likely changed cluster state
	for _, cmd := range m.recentCommands {
		cmdLower := strings.ToLower(cmd)
		if strings.Contains(cmdLower, "apply") ||
		   strings.Contains(cmdLower, "create") ||
		   strings.Contains(cmdLower, "delete") ||
		   strings.Contains(cmdLower, "patch") ||
		   strings.Contains(cmdLower, "scale") ||
		   strings.Contains(cmdLower, "install") ||
		   strings.Contains(cmdLower, "upgrade") {
			return true
		}
	}
	return false
}

func (m *model) trackUserInput() {
	m.lastUserInput = time.Now()
}

func (m *model) trackCommand(command string) {
	// Add command to recent commands
	m.recentCommands = append(m.recentCommands, command)

	// Keep only recent commands (rolling window)
	if len(m.recentCommands) > m.maxRecentCommands {
		m.recentCommands = m.recentCommands[1:]
	}
}

// handleAsyncIntelligenceResult processes async intelligence results
func (m *model) handleAsyncIntelligenceResult(result engine.IntelligenceResult) {
	// Remove from pending work
	delete(m.pendingIntelligenceWork, result.ID)

	if !result.Success {
		// Handle error case
		if result.Error != nil {
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: fmt.Sprintf("⚠️ Intelligence analysis failed: %v", result.Error),
			})
		}
		return
	}

	// Process successful result
	switch result.Type {
	case WorkTypeAnalysis:
		if session, ok := result.Data.(*AnalysisSession); ok {
			// Update intelligence UI with the new analysis
			if m.intelligentUI != nil {
				m.riskLevel = m.intelligentUI.GetCurrentRiskLevel()
				m.lastIntelligenceUpdate = time.Now()

				// Show suggestions if confidence is low
				if session.Confidence < 0.7 {
					suggestions := m.intelligentUI.FormatSuggestions()
					if suggestions != "" {
						m.messages = append(m.messages, message{
							sender:  systemSender,
							content: "🤖 AI Analysis complete - suggestions available (F1-F3)",
						})
					}
				}
			}
		}
	}

	// Refresh UI
	m.chatViewport.SetContent(m.renderMessages())
	m.chatViewport.GotoBottom()
}

// handleStreamingIntelligenceUpdate processes streaming intelligence updates
func (m *model) handleStreamingIntelligenceUpdate(update IntelligenceUpdate) {
	if m.intelligentUI == nil {
		return
	}

	switch update.Type {
	case UpdateTypeRiskChange:
		if riskData, ok := update.Data.(RiskLevel); ok {
			m.riskLevel = riskData
			// Publish risk change notification if significant
			if riskData.Level == "high" || riskData.Level == "critical" {
				m.messages = append(m.messages, message{
					sender:  systemSender,
					content: fmt.Sprintf("⚠️ Risk level changed to %s: %s",
						strings.ToUpper(riskData.Level),
						strings.Join(riskData.Factors, ", ")),
				})
			}
		}

	case UpdateTypeContextChange:
		if contextData, ok := update.Data.(*KubeContextSummary); ok {
			m.currentContext = contextData
			m.ctxName = contextData.Context
			m.namespace = contextData.Namespace

			// Invalidate related cache entries
			if update.Invalidates != nil {
				for _, cacheKey := range update.Invalidates {
					Intelligence.InvalidateCache(cacheKey)
				}
			}
		}

	case UpdateTypeHealthChange:
		if healthData, ok := update.Data.(string); ok {
			oldHealth := m.clusterHealth
			m.clusterHealth = healthData

			// Notify if health degraded
			if (oldHealth == "healthy" && healthData != "healthy") ||
			   (oldHealth != "critical" && healthData == "critical") {
				icon := "⚠️"
				if healthData == "critical" {
					icon = "🚨"
				}
				m.messages = append(m.messages, message{
					sender:  systemSender,
					content: fmt.Sprintf("%s Cluster health changed: %s → %s",
						icon, oldHealth, healthData),
				})
			}
		}

	case UpdateTypeNewSuggestion:
		if suggestion, ok := update.Data.(AISuggestion); ok {
			// Add new suggestion to UI
			m.intelligentUI.suggestions = append(m.intelligentUI.suggestions, suggestion)

			// Notify user of high-confidence suggestions
			if suggestion.Confidence > 0.8 && update.Priority >= 7 {
				m.messages = append(m.messages, message{
					sender:  systemSender,
					content: fmt.Sprintf("💡 High-confidence suggestion: %s (%.0f%%) [%s]",
						suggestion.Title,
						suggestion.Confidence*100,
						suggestion.Hotkey),
				})
			}
		}

	case UpdateTypePrediction:
		// Handle predictive intelligence updates
		if update.Priority >= 8 {
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: "🔮 Predictive intelligence updated with new patterns",
			})
		}

	case UpdateTypeOptimization:
		// Handle optimization recommendations
		if update.Priority >= 7 {
			m.messages = append(m.messages, message{
				sender:  systemSender,
				content: "⚡ New optimization recommendations available",
			})
		}
	}

	// Update activity tracking
	if m.streamingManager != nil {
		activity := m.calculateCurrentActivity()
		typingSpeed := m.calculateTypingSpeed()
		m.streamingManager.UpdateUserActivity(activity, typingSpeed)
	}

	// Refresh UI components
	m.updateIntelligenceUI()
}

// refreshIntelligenceAsync uses async intelligence processing when possible
func (m *model) refreshIntelligenceAsync() {
	if m.intelligentUI == nil {
		return
	}

	// Get current context
	context, err := BuildContextSummary()
	if err == nil {
		m.currentContext = context
		m.ctxName = context.Context
		m.namespace = context.Namespace

		// Update cluster health
		if len(context.PodProblemCounts) == 0 {
			m.clusterHealth = "healthy"
		} else if len(context.PodProblemCounts) > 5 {
			m.clusterHealth = "critical"
		} else {
			m.clusterHealth = "warning"
		}
	}

	// Get current input
	input := strings.TrimSpace(m.textarea.Value())
	if input == "" {
		return
	}

	// Use async processing if enabled
	if m.asyncIntelligenceEnabled && Intelligence.processor != nil {
		callback := func(result engine.IntelligenceResult) {
			// This will be handled by the async result message
		}

		// Determine priority based on user activity
		if m.isUserTyping() {
			workID := Intelligence.AnalyzeIntelligentlyHighPriority(input, m.currentContext, callback)
			m.pendingIntelligenceWork[workID] = true
		} else {
			workID := Intelligence.AnalyzeIntelligentlyAsync(input, m.currentContext, callback)
			m.pendingIntelligenceWork[workID] = true
		}
	} else {
		// Fallback to sync processing
		err := m.intelligentUI.UpdateIntelligence(input, m.currentContext)
		if err == nil {
			m.riskLevel = m.intelligentUI.GetCurrentRiskLevel()
			m.lastIntelligenceUpdate = time.Now()
		}
	}
}

// Helper methods for streaming intelligence

func (m *model) calculateCurrentActivity() ActivityLevel {
	if m.textarea.Focused() {
		timeSinceInput := time.Since(m.lastUserInput)
		if timeSinceInput < 2*time.Second {
			return ActivityIntense
		} else if timeSinceInput < 5*time.Second {
			return ActivityHigh
		} else if timeSinceInput < 15*time.Second {
			return ActivityModerate
		} else if timeSinceInput < 60*time.Second {
			return ActivityLow
		}
	}
	return ActivityIdle
}

func (m *model) calculateTypingSpeed() float64 {
	// Calculate characters per minute based on recent typing
	if len(m.textarea.Value()) == 0 {
		return 0.0
	}

	timeSinceInput := time.Since(m.lastUserInput)
	if timeSinceInput == 0 {
		return 0.0
	}

	charactersTyped := float64(len(m.textarea.Value()))
	minutes := timeSinceInput.Minutes()
	if minutes == 0 {
		return charactersTyped * 60.0 // Assume 1 second = 1/60 minute
	}

	return charactersTyped / minutes
}

func (m *model) updateIntelligenceUI() {
	// Force refresh of intelligence UI components
	m.chatViewport.SetContent(m.renderMessages())
	m.outputViewport.SetContent(m.formatRightPane())
}

func (m *model) publishIntelligenceUpdate(updateType IntelligenceUpdateType, data interface{}, priority int) {
	if m.streamingManager == nil {
		return
	}

	update := IntelligenceUpdate{
		ID:        fmt.Sprintf("%s_%d", updateType, time.Now().UnixNano()),
		Type:      updateType,
		Priority:  priority,
		Source:    "tui",
		Data:      data,
		Context:   m.currentContext,
		Confidence: 0.8, // Default confidence
		TTL:       5 * time.Minute,
	}

	m.streamingManager.PublishUpdate(update)
}
