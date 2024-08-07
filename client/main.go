package main

import (
	"fmt"
	"gophkeep/client/internal/communication"
	"gophkeep/client/internal/gophmodel"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	width        int
	height       int
	currentStage string
	errorMessage string
	globalState  string
	loginInfo    gophmodel.SimpleAccountData
	textInput    textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 25
	ti.Width = 20

	return model{
		textInput:    ti,
		errorMessage: "",
	}
}

func (m model) Init() tea.Cmd {
	return startAppCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentStage {
	case "PingFail":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				return m, handlePingServer
			}
		}
	case "SignInChoise":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "l":
				m.currentStage = "LoginRegisterInputs"
				m.globalState = "login"
				m.textInput.Placeholder = "Enter your login here"
				return m, nil
			case "r":
				m.currentStage = "LoginRegisterInputs"
				m.globalState = "register"
				m.textInput.Placeholder = "Enter your new login here"
				return m, nil
			}
		}
	case "LoginRegisterInputs":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "PasswordInput"
				m.loginInfo.Login = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
	case "PasswordInput":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "Auth"
				m.loginInfo.Password = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
		
	case "Auth":
		return m, handlePingServer
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case startAppMsg:
		{
			m.currentStage = "PingServer"
			return m, handlePingServer
		}
	case stageCompleteMsg:
		{
			m.currentStage = msg.NextStageNameKey
			m.errorMessage = msg.ErrorMessage
		}
	}

	return m, nil
}

func (m model) View() string {
	var s string
	switch m.currentStage {
	case "PingServer":
		s = "Connecting to server"
	case "SignInChoise":
		s = "type 'l' or 'r' to login or register"
	case "PingFail":
		s = m.errorMessage + "\n Could not connect to the server" +
			"\n Press Enter to retry"

	case "LoginRegisterInputs":
		return fmt.Sprintf(
			"Input your login: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "PasswordInput":
		m.textInput.Placeholder = "Password"
		m.textInput.EchoMode = textinput.EchoPassword
		m.textInput.EchoCharacter = '*'
		return fmt.Sprintf(
			"Input your password: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"

	case "Auth":
		s = "Authentification"
	}

	return "\n" + s + "\n\n"
}

type startAppMsg struct{}

func startAppCmd() tea.Msg {
	return startAppMsg{}
}

type stageCompleteMsg struct {
	NextStageNameKey string
	ErrorMessage     string
}

type Stage struct {
	Name   string
	Action func() error
	Error  error
	Reset  func() error
}

var stages = map[string]Stage{
	"PingServer": {
		Name: "PingServer",
		Action: func() error {
			return nil
		},
	},

	"PingFail": {
		Name: "PingFail",
		Action: func() error {
			return nil
		},
	},
	"SignInChoise": {
		Name: "SignInChoise",
		Action: func() error {
			return nil
		},
	},
	"LoginRegisterInputs": {
		Name: "LoginRegisterInputs",
		Action: func() error {
			return nil
		},
	},
	"AuthComplete": {
		Name: "AuthComplete",
		Action: func() error {
			return nil
		},
	},
	"AuthError": {
		Name: "AuthError",
		Action: func() error {
			return nil
		},
	},
}

func handleLogin(m model) tea.Msg{
	status, err := communication.LoginHandle()

}

func handlePingServer() tea.Msg {
	status, err := communication.PingServerHandle()
	var msg stageCompleteMsg
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "PingFail"
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "status is not 200, it is: " + string(status)
		msg.NextStageNameKey = "PingFail"
		return msg
	}
	if status == 200 {
		msg.ErrorMessage = ""
		msg.NextStageNameKey = "SignInChoise"
		return msg
	}
	return msg
}

func main() {

	f, err := tea.LogToFile("debug.txt", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
