package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const githubDiffURL = "https://github.com/k3s-io/k3s/compare"

type editorFinishedMsg struct{ err error }

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	// helpStyle           = blurredStyle.Copy()
	titleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#E678E6")).Bold(true)
	addDiffStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	removeDiffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

	headerStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type model struct {
	focusIndex    int
	tagsSubmitted bool
	// diff          []string
	inputs        []textinput.Model
	viewport      viewport.Model
	viewportReady bool
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel() model {
	m := model{
		inputs:        make([]textinput.Model, 2),
		tagsSubmitted: false,
		viewportReady: false,
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 20
		switch i {
		case 0:
			t.Placeholder = "New tag"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Old tag"
		}
		m.inputs[i] = t
	}
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.viewportReady {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.HighPerformanceRendering = false
			m.viewportReady = true
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			if !m.tagsSubmitted {
				s := msg.String()
				if s == "enter" && m.focusIndex == len(m.inputs) {
					diff := styledDiff(m.inputs[0].Value(), m.inputs[1].Value())
					fmt.Print(diff)
					m.viewport.SetContent(diff)
					m.tagsSubmitted = true
				}
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs)
				}

				inputCmds := make([]tea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						inputCmds[i] = m.inputs[i].Focus()
						m.inputs[i].PromptStyle = focusedStyle
						m.inputs[i].TextStyle = focusedStyle
						continue
					}
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = noStyle
					m.inputs[i].TextStyle = noStyle
				}

				cmds = append(cmds, inputCmds...)
			} else {
				if s := msg.String(); s == "enter" {
					cmd = tea.ExitAltScreen
					return m, tea.Batch(cmd, openEditor())
				}
			}
		}
	}
	if !m.tagsSubmitted {
		cmds = append(cmds, m.updateInputs(msg))
	}
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func styledDiff(newTag, oldTag string) string {
	diff, err := agentAndServerArgsDiff(githubDiffURL, newTag, oldTag)
	if err != nil {
		log.Fatal(err)
	}
	diffContent := []string{}
	for _, line := range diff {
		if line[0] == '+' && !strings.Contains(line, "@@") {
			diffContent = append(diffContent, addDiffStyle.Render(line))
		} else if line[0] == '-' {
			diffContent = append(diffContent, removeDiffStyle.Render(line))
		} else {
			diffContent = append(diffContent, noStyle.Render(line))
		}
	}
	return strings.Join(diffContent, "\n")
}

func (m model) View() string {
	if !m.tagsSubmitted {
		return inputView(m)
	}
	// return inputView(m)
	return diffView(m)
}

func inputView(m model) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Input tags to get the diff for server and agent args in k3s"))
	b.WriteString("\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func diffView(m model) string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := headerStyle.Render("Diff Viewer")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}
func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	help := infoStyle.Render("press enter to continue and open your editor")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)-lipgloss.Width(help)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, help, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func openEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	// wd, err := os.Getwd()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	wd := "/Users/pedrosuse/code/tashima42/kontainer-driver-metadata"
	wd = filepath.Join(wd, "channels.yaml")
	command := editor /*+ " " + wd*/
	// fmt.Println(command)
	c := exec.Command(command) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
