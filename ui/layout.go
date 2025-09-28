package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// LayoutMode represents different layout configurations
type LayoutMode int

const (
	LayoutThreePane LayoutMode = iota
	LayoutVerticalSplit
	LayoutHorizontalSplit
	LayoutChatOnly
)

// PaneType represents different pane types in the layout
type PaneType int

const (
	PaneChat PaneType = iota
	PanePreview
	PaneOutput
)

// PaneConfig holds configuration for a single pane
type PaneConfig struct {
	Width   int
	Height  int
	X       int
	Y       int
	Visible bool
	Title   string
}

// LayoutConfig holds the complete layout configuration
type LayoutConfig struct {
	Mode    LayoutMode
	Chat    PaneConfig
	Preview PaneConfig
	Output  PaneConfig
	Window  struct {
		Width  int
		Height int
	}
	ContentArea struct {
		Width  int
		Height int
	}
}

// LayoutManager handles layout calculations and transitions
type LayoutManager struct {
	config        LayoutConfig
	theme         *Theme
	minPaneWidth  int
	minPaneHeight int
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager(theme *Theme) *LayoutManager {
	return &LayoutManager{
		theme:         theme,
		minPaneWidth:  30,
		minPaneHeight: 6,
		config: LayoutConfig{
			Mode: LayoutThreePane,
		},
	}
}

// Next returns the next layout mode in the cycle
func (l LayoutMode) Next() LayoutMode {
	switch l {
	case LayoutThreePane:
		return LayoutVerticalSplit
	case LayoutVerticalSplit:
		return LayoutHorizontalSplit
	case LayoutHorizontalSplit:
		return LayoutChatOnly
	default:
		return LayoutThreePane
	}
}

// String returns a human-readable name for the layout mode
func (l LayoutMode) String() string {
	switch l {
	case LayoutThreePane:
		return "Three Pane"
	case LayoutVerticalSplit:
		return "Vertical Split"
	case LayoutHorizontalSplit:
		return "Horizontal Split"
	case LayoutChatOnly:
		return "Chat Only"
	default:
		return "Unknown"
	}
}

// SetWindowSize updates the window dimensions and recalculates layout
func (lm *LayoutManager) SetWindowSize(width, height int) {
	lm.config.Window.Width = width
	lm.config.Window.Height = height
	lm.calculateContentArea()
	lm.calculateLayout()
}

// CycleLayout switches to the next layout mode
func (lm *LayoutManager) CycleLayout() {
	lm.config.Mode = lm.config.Mode.Next()
	lm.calculateLayout()
}

// GetLayoutMode returns the current layout mode
func (lm *LayoutManager) GetLayoutMode() LayoutMode {
	return lm.config.Mode
}

// GetConfig returns the current layout configuration
func (lm *LayoutManager) GetConfig() LayoutConfig {
	return lm.config
}

// calculateContentArea determines the available content area
func (lm *LayoutManager) calculateContentArea() {
	// Reserve space for header, input, status bar, etc.
	headerHeight := 2 // Header and spacing
	inputHeight := 5  // Input area with borders
	statusHeight := 1 // Status bar
	footerHeight := 2 // Footer and spacing

	reservedHeight := headerHeight + inputHeight + statusHeight + footerHeight

	lm.config.ContentArea.Width = lm.config.Window.Width - 6 // Margins
	if lm.config.ContentArea.Width < 60 {
		lm.config.ContentArea.Width = 60
	}

	lm.config.ContentArea.Height = lm.config.Window.Height - reservedHeight
	if lm.config.ContentArea.Height < 15 {
		lm.config.ContentArea.Height = 15
	}
}

// calculateLayout calculates pane positions and sizes based on current mode
func (lm *LayoutManager) calculateLayout() {
	switch lm.config.Mode {
	case LayoutThreePane:
		lm.calculateThreePaneLayout()
	case LayoutVerticalSplit:
		lm.calculateVerticalSplitLayout()
	case LayoutHorizontalSplit:
		lm.calculateHorizontalSplitLayout()
	case LayoutChatOnly:
		lm.calculateChatOnlyLayout()
	}
}

// calculateThreePaneLayout implements the three-pane layout
func (lm *LayoutManager) calculateThreePaneLayout() {
	contentWidth := lm.config.ContentArea.Width
	contentHeight := lm.config.ContentArea.Height

	// Chat takes 55% of width, remaining split between preview and output
	chatWidth := int(float64(contentWidth) * 0.55)
	if chatWidth < lm.minPaneWidth {
		chatWidth = lm.minPaneWidth
	}

	rightWidth := contentWidth - chatWidth - 1 // 1 for separator
	if rightWidth < lm.minPaneWidth {
		rightWidth = lm.minPaneWidth
		chatWidth = contentWidth - rightWidth - 1
	}

	// Right side split vertically between preview and output
	previewHeight := contentHeight / 2
	if previewHeight < lm.minPaneHeight {
		previewHeight = lm.minPaneHeight
	}

	outputHeight := contentHeight - previewHeight - 1 // 1 for separator
	if outputHeight < lm.minPaneHeight {
		outputHeight = lm.minPaneHeight
		previewHeight = contentHeight - outputHeight - 1
	}

	lm.config.Chat = PaneConfig{
		Width:   chatWidth,
		Height:  contentHeight,
		X:       0,
		Y:       0,
		Visible: true,
		Title:   "Chat",
	}

	lm.config.Preview = PaneConfig{
		Width:   rightWidth,
		Height:  previewHeight,
		X:       chatWidth + 1,
		Y:       0,
		Visible: true,
		Title:   "Preview",
	}

	lm.config.Output = PaneConfig{
		Width:   rightWidth,
		Height:  outputHeight,
		X:       chatWidth + 1,
		Y:       previewHeight + 1,
		Visible: true,
		Title:   "Output / Logs",
	}
}

// calculateVerticalSplitLayout implements the vertical split layout
func (lm *LayoutManager) calculateVerticalSplitLayout() {
	contentWidth := lm.config.ContentArea.Width
	contentHeight := lm.config.ContentArea.Height

	// Chat takes 60% of width, right side combines preview and output
	chatWidth := int(float64(contentWidth) * 0.6)
	if chatWidth < lm.minPaneWidth {
		chatWidth = lm.minPaneWidth
	}

	rightWidth := contentWidth - chatWidth - 1
	if rightWidth < lm.minPaneWidth {
		rightWidth = lm.minPaneWidth
		chatWidth = contentWidth - rightWidth - 1
	}

	lm.config.Chat = PaneConfig{
		Width:   chatWidth,
		Height:  contentHeight,
		X:       0,
		Y:       0,
		Visible: true,
		Title:   "Chat",
	}

	lm.config.Preview = PaneConfig{
		Width:   rightWidth,
		Height:  contentHeight,
		X:       chatWidth + 1,
		Y:       0,
		Visible: true,
		Title:   "Preview & Output",
	}

	lm.config.Output = PaneConfig{
		Width:   0,
		Height:  0,
		X:       0,
		Y:       0,
		Visible: false,
		Title:   "",
	}
}

// calculateHorizontalSplitLayout implements the horizontal split layout
func (lm *LayoutManager) calculateHorizontalSplitLayout() {
	contentWidth := lm.config.ContentArea.Width
	contentHeight := lm.config.ContentArea.Height

	// Chat takes 55% of height, bottom combines preview and output
	chatHeight := int(float64(contentHeight) * 0.55)
	if chatHeight < lm.minPaneHeight {
		chatHeight = lm.minPaneHeight
	}

	bottomHeight := contentHeight - chatHeight - 1
	if bottomHeight < lm.minPaneHeight {
		bottomHeight = lm.minPaneHeight
		chatHeight = contentHeight - bottomHeight - 1
	}

	lm.config.Chat = PaneConfig{
		Width:   contentWidth,
		Height:  chatHeight,
		X:       0,
		Y:       0,
		Visible: true,
		Title:   "Chat",
	}

	lm.config.Preview = PaneConfig{
		Width:   contentWidth,
		Height:  bottomHeight,
		X:       0,
		Y:       chatHeight + 1,
		Visible: true,
		Title:   "Preview & Output",
	}

	lm.config.Output = PaneConfig{
		Width:   0,
		Height:  0,
		X:       0,
		Y:       0,
		Visible: false,
		Title:   "",
	}
}

// calculateChatOnlyLayout implements the chat-only layout
func (lm *LayoutManager) calculateChatOnlyLayout() {
	contentWidth := lm.config.ContentArea.Width
	contentHeight := lm.config.ContentArea.Height

	lm.config.Chat = PaneConfig{
		Width:   contentWidth,
		Height:  contentHeight,
		X:       0,
		Y:       0,
		Visible: true,
		Title:   "Chat",
	}

	lm.config.Preview = PaneConfig{
		Width:   0,
		Height:  0,
		X:       0,
		Y:       0,
		Visible: false,
		Title:   "",
	}

	lm.config.Output = PaneConfig{
		Width:   0,
		Height:  0,
		X:       0,
		Y:       0,
		Visible: false,
		Title:   "",
	}
}

// RenderLayout returns the complete layout view
func (lm *LayoutManager) RenderLayout(chatContent, previewContent, outputContent string) string {
	theme := CurrentTheme()

	var sections []string

	switch lm.config.Mode {
	case LayoutThreePane:
		sections = append(sections, lm.renderThreePaneLayout(chatContent, previewContent, outputContent, theme))
	case LayoutVerticalSplit:
		sections = append(sections, lm.renderVerticalSplitLayout(chatContent, previewContent, outputContent, theme))
	case LayoutHorizontalSplit:
		sections = append(sections, lm.renderHorizontalSplitLayout(chatContent, previewContent, outputContent, theme))
	case LayoutChatOnly:
		sections = append(sections, lm.renderChatOnlyLayout(chatContent, theme))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderThreePaneLayout renders the three-pane layout
func (lm *LayoutManager) renderThreePaneLayout(chatContent, previewContent, outputContent string, theme *Theme) string {
	chatPane := lm.renderPane("Chat", chatContent, lm.config.Chat, theme)
	previewPane := lm.renderPane(lm.config.Preview.Title, previewContent, lm.config.Preview, theme)
	outputPane := lm.renderPane(lm.config.Output.Title, outputContent, lm.config.Output, theme)

	rightColumn := lipgloss.JoinVertical(lipgloss.Left, previewPane, outputPane)
	return lipgloss.JoinHorizontal(lipgloss.Top, chatPane, rightColumn)
}

// renderVerticalSplitLayout renders the vertical split layout
func (lm *LayoutManager) renderVerticalSplitLayout(chatContent, previewContent, outputContent string, theme *Theme) string {
	chatPane := lm.renderPane("Chat", chatContent, lm.config.Chat, theme)

	// Combine preview and output content
	combinedContent := previewContent
	if outputContent != "" {
		if previewContent != "" {
			combinedContent += "\n\n" + outputContent
		} else {
			combinedContent = outputContent
		}
	}

	rightPane := lm.renderPane(lm.config.Preview.Title, combinedContent, lm.config.Preview, theme)
	return lipgloss.JoinHorizontal(lipgloss.Top, chatPane, rightPane)
}

// renderHorizontalSplitLayout renders the horizontal split layout
func (lm *LayoutManager) renderHorizontalSplitLayout(chatContent, previewContent, outputContent string, theme *Theme) string {
	chatPane := lm.renderPane("Chat", chatContent, lm.config.Chat, theme)

	// Combine preview and output content
	combinedContent := previewContent
	if outputContent != "" {
		if previewContent != "" {
			combinedContent += "\n\n" + outputContent
		} else {
			combinedContent = outputContent
		}
	}

	bottomPane := lm.renderPane(lm.config.Preview.Title, combinedContent, lm.config.Preview, theme)
	return lipgloss.JoinVertical(lipgloss.Left, chatPane, bottomPane)
}

// renderChatOnlyLayout renders the chat-only layout
func (lm *LayoutManager) renderChatOnlyLayout(chatContent string, theme *Theme) string {
	return lm.renderPane("Chat", chatContent, lm.config.Chat, theme)
}

// renderPane renders a single pane with title and content
func (lm *LayoutManager) renderPane(title, content string, config PaneConfig, theme *Theme) string {
	if !config.Visible {
		return ""
	}

	if content == "" {
		content = "──"
	}

	titleStyle := theme.HighlightStyle("keyword").Bold(true)
	titleBlock := titleStyle.Render(title)

	paneStyle := theme.PaneStyle().
		Width(config.Width).
		Height(config.Height)

	contentBlock := paneStyle.Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, titleBlock, contentBlock)
}

// GetPaneConfig returns the configuration for a specific pane type
func (lm *LayoutManager) GetPaneConfig(paneType PaneType) PaneConfig {
	switch paneType {
	case PaneChat:
		return lm.config.Chat
	case PanePreview:
		return lm.config.Preview
	case PaneOutput:
		return lm.config.Output
	default:
		return PaneConfig{}
	}
}

// SetPreviewTitle updates the preview pane title
func (lm *LayoutManager) SetPreviewTitle(title string) {
	lm.config.Preview.Title = title
}

// SetOutputTitle updates the output pane title
func (lm *LayoutManager) SetOutputTitle(title string) {
	lm.config.Output.Title = title
}
