package main

import (
	"encoding/json"
	"fmt"
	"gophkeep/client/internal/communication"
	gophmodel "gophkeep/internal/model"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	width            int
	height           int
	currentStage     string
	errorMessage     string
	globalState      string
	targetObjectName string

	loginInfo    gophmodel.SimpleAccountData
	userMetadata []gophmodel.Metadata
	textInput    textinput.Model
	clientEnv    *communication.ClientEnv

	simpleMetadata       gophmodel.SimpleMetadata
	loginAndPasswordData gophmodel.LoginAndPasswordData
	cardData             gophmodel.CardData
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 20

	return model{
		textInput:    ti,
		errorMessage: "",
		clientEnv:    &communication.ClientEnv{},
	}
}

func (m model) Init() tea.Cmd {
	return startAppCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
			return m, handlePingServer(m)
		}
	case stageCompleteMsg:
		{
			m.currentStage = msg.NextStageNameKey
			m.errorMessage = msg.ErrorMessage
		}
	}

	switch m.currentStage {
	case "PingFail":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				return m, handlePingServer(m)
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
			case tea.KeyRunes, tea.KeyBackspace:
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
			case tea.KeyRunes, tea.KeyBackspace:
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
		if m.globalState == "login" {
			return m, handleLogin(m)
		}

		if m.globalState == "register" {
			return m, handleRegister(m)
		}
	case "AuthFailed":
		m.currentStage = "SignInChoise"
		return m, nil
	case "AuthSuccess":
		m.currentStage = "Sync"
		return m, nil
	case "Sync":
		return m, handleSync(m)
	case "MainMenu":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				msg := handleMainMenuCommand(m)
				m.textInput.SetValue("")
				return m, msg
			}
		}
	case "Write":
		m.currentStage = "WriteName"
		return m, nil
	case "WriteName":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteDescription"
				m.simpleMetadata.Name = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
		return m, nil
	case "WriteDescription":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "SelectDataType"
				m.simpleMetadata.Description = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
		return m, nil
	case "SelectDataType":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "1":
				m.currentStage = "WriteLogin"
				m.simpleMetadata.DataType = "passwords"
				return m, nil
			case "2":
				m.currentStage = "WriteCardData"
				m.simpleMetadata.DataType = "cards"
				return m, nil
			}
		}
	case "WriteLogin":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WritePassword"
				m.loginAndPasswordData.Login = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
	case "WritePassword":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteLoginAndPasswordToServer"
				m.loginAndPasswordData.Password = m.textInput.Value()
				m.textInput.SetValue("")
				return m, nil
			}
		}
	case "WriteLoginAndPasswordToServer":
		return m, handleWriteLoginAndPassword(m)
	case "DataSaved":
		m.currentStage = "MainMenu"
		return m, nil
	case "List":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.currentStage = "MainMenu"
				return m, nil
			}
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
		if len(m.errorMessage) != 0 {
			s = m.errorMessage + "\n" + s
		}
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
	case "Sync":
		s = "Loading data from server"
	case "MainMenu":
		errorMessage := ""
		if len(m.errorMessage) != 0 {
			errorMessage = m.errorMessage + "\n\n"
		}
		m.textInput.Placeholder = "Type command here"
		s = errorMessage + "Menu\n\n" + "type:\n\nread <name> to read your saved data" +
			"\n\nwrite to add new data" +
			"\n\nlist to view all names and descriptions of your data" +
			"\n\ndelete <name> to delete data" +
			"\n\nedit <name> to edit data \n\n" + m.textInput.View()
	case "WriteName":
		return fmt.Sprintf(
			"Input name of the data: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteDescription":
		return fmt.Sprintf(
			"Input description of the data: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "SelectDataType":
		s = "press '1' to add login and password data or press '2' to add card data"
	case "WriteLogin":
		return fmt.Sprintf(
			"Input your login: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WritePassword":
		m.textInput.Placeholder = "Password"
		m.textInput.EchoMode = textinput.EchoPassword
		m.textInput.EchoCharacter = '*'
		return fmt.Sprintf(
			"Input your password: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "DataSaved":
		s = "Data added"
	case "List":
		s = drawList(m)
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

func handleLogin(m model) tea.Cmd {
	return func() tea.Msg {
		var msg stageCompleteMsg
		status, err := m.clientEnv.LoginHandle(m.loginInfo)
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "AuthFailed"
			return msg
		}
		if status == 401 {
			msg.ErrorMessage = "no such login and password pair found"
			msg.NextStageNameKey = "AuthFailed"
			return msg
		}
		if status == 200 {
			msg.ErrorMessage = ""
			msg.NextStageNameKey = "AuthSuccess"
			return msg
		}
		msg.ErrorMessage = "server error, unexpected status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "AuthFailed"
		return msg
	}
}

func handleRegister(m model) tea.Cmd {
	return func() tea.Msg {
		var msg stageCompleteMsg
		status, err := m.clientEnv.RegisterHandle(m.loginInfo)
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "AuthFailed"
			return msg
		}
		if status == 409 {
			msg.ErrorMessage = "login alredy in use"
			msg.NextStageNameKey = "AuthFailed"
			return msg
		}
		if status == 200 {
			msg.ErrorMessage = ""
			msg.NextStageNameKey = "AuthSuccess"
			return msg
		}
		msg.ErrorMessage = "server error, unexpected status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "AuthFailed"
		return msg
	}
}

func handlePingServer(m model) tea.Cmd {
	return func() tea.Msg {
		status, err := m.clientEnv.PingServerHandle()
		var msg stageCompleteMsg
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "PingFail"
			return msg
		}
		if status != 200 {
			msg.ErrorMessage = "status is not 200, it is: " + fmt.Sprint(status)
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
}

func handleSync(m model) tea.Cmd {
	return func() tea.Msg {
		status, userMetadata, err := m.clientEnv.SyncHandle()
		m.userMetadata = userMetadata

		var msg stageCompleteMsg
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "SyncFail"
			return msg
		}
		if status == 204 {
			msg.ErrorMessage = ""
			msg.NextStageNameKey = "MainMenu"
			return msg
		}
		if status == 200 {
			msg.ErrorMessage = ""
			msg.NextStageNameKey = "MainMenu"
			return msg
		}
		return msg
	}
}

func handleMainMenuCommand(m model) tea.Cmd {
	return func() tea.Msg {
		var msg stageCompleteMsg
		command := m.textInput.Value()
		commandSlice := strings.Fields(command)
		switch len(commandSlice) {
		case 0:
			msg.ErrorMessage = "command didn't have any words"
			msg.NextStageNameKey = "MainMenu"
		case 1:
			switch commandSlice[0] {
			case "write":
				msg.ErrorMessage = ""
				msg.NextStageNameKey = "Write"
			case "list":
				msg.ErrorMessage = ""
				msg.NextStageNameKey = "List"
			default:
				msg.ErrorMessage = "Unknown command"
				msg.NextStageNameKey = "MainMenu"
			}
		case 2:
			switch commandSlice[0] {
			case "read":
				msg.ErrorMessage = ""
				msg.NextStageNameKey = "Read"
			case "delete":
				msg.ErrorMessage = ""
				msg.NextStageNameKey = "Delete"
			case "edit":
				msg.ErrorMessage = ""
				msg.NextStageNameKey = "Edit"
			default:
				msg.ErrorMessage = "Unknown command"
				msg.NextStageNameKey = "MainMenu"
			}
			m.targetObjectName = commandSlice[1]
		default:
			msg.ErrorMessage = "Unknown command"
			msg.NextStageNameKey = "MainMenu"
		}
		return msg
	}
}

func handleWriteLoginAndPassword(m model) tea.Cmd {
	return func() tea.Msg {
		var msg stageCompleteMsg

		var data []byte

		switch m.simpleMetadata.DataType {
		case "passwords":
			bytes, err := json.Marshal(m.loginAndPasswordData)
			if err != nil {
				msg.ErrorMessage = err.Error()
				msg.NextStageNameKey = "MainMenu"
			}
			data = bytes
		}

		status, metadata, err := m.clientEnv.WriteHandle(m.simpleMetadata, data)
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "MainMenu"
			return msg
		}
		if status == 200 {
			msg.ErrorMessage = ""
			msg.NextStageNameKey = "DataSaved"
			m.userMetadata = append(m.userMetadata, metadata)
			return msg
		}
		if status != 200 {
			msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
			msg.NextStageNameKey = "MainMenu"
			return msg
		}

		return msg
	}
}

func drawList(m model) string {
	var sb strings.Builder
	sb.WriteString("Name, Description, Data Type, Changed, Created")
	count := len(m.userMetadata)
	for i := 0; i < count; i++ {
		sb.WriteString(fmt.Sprintf("Name: %s , Description: %s , Data Type: %s , Changed: %s , Created: %s\n\n ,",
			m.userMetadata[i].Name,
			m.userMetadata[i].Description,
			m.userMetadata[i].DataType,
			m.userMetadata[i].Changed,
			m.userMetadata[i].Created,
		))
	}
	return sb.String()
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
