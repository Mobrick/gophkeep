package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	width  int
	height int
}

func New()

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}
	return "xdd"
}

func main() {

	f, err := tea.LogToFile("debug.txt", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
