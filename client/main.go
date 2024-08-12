package main

import (
	"encoding/json"
	"fmt"
	"gophkeep/client/internal/communication"
	gophmodel "gophkeep/internal/model"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	width            int
	height           int
	currentStage     string
	errorMessage     string
	globalState      string
	targetObjectName *string
	outputData       *string

	loginInfo    gophmodel.SimpleAccountData
	userMetadata *[]gophmodel.Metadata
	textInput    textinput.Model
	clientEnv    *communication.ClientEnv

	targetObjectMetadata gophmodel.Metadata
	targetObjectIndex    int
	newMetadata          gophmodel.SimpleMetadata
	loginAndPasswordData *gophmodel.LoginAndPasswordData
	cardData             *gophmodel.CardData
}

type tickMsg time.Time

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 20

	objectName := ""
	outputData := ""

	return model{
		textInput:            ti,
		errorMessage:         "",
		clientEnv:            &communication.ClientEnv{},
		userMetadata:         &[]gophmodel.Metadata{},
		targetObjectName:     &objectName,
		outputData:           &outputData,
		loginAndPasswordData: &gophmodel.LoginAndPasswordData{},
		cardData:             &gophmodel.CardData{},
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
				return m, cmd
			case "r":
				m.currentStage = "LoginRegisterInputs"
				m.globalState = "register"
				m.textInput.Placeholder = "Enter your new login here"
				return m, cmd
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
				return m, cmd
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
				return m, cmd
			}
		}

	case "Auth":
		if m.globalState == "login" {
			msg := handleLogin(m)
			m.currentStage = msg.NextStageNameKey
			m.errorMessage = msg.ErrorMessage
			return m, cmd
		}

		if m.globalState == "register" {
			msg := handleRegister(m)
			m.currentStage = msg.NextStageNameKey
			m.errorMessage = msg.ErrorMessage
			return m, cmd
		}
	case "AuthFailed":
		m.currentStage = "SignInChoise"
		return m, cmd
	case "Sync":
		msg := handleSync(m)
		m.currentStage = msg.NextStageNameKey
		m.errorMessage = msg.ErrorMessage
		return m, tick()
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
				m.newMetadata.Name = m.textInput.Value()
				if nameAlreadyExists(m, m.newMetadata.Name) {
					m.errorMessage = "you already use that data name"
					m.currentStage = "MainMenu"
				} else {
					m.currentStage = "WriteDescription"
				}
				m.textInput.SetValue("")
				return m, cmd
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
				m.newMetadata.Description = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
		return m, nil
	case "SelectDataType":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "1":
				m.currentStage = "WriteLogin"
				m.newMetadata.DataType = "passwords"
				return m, cmd
			case "2":
				m.currentStage = "WriteNumber"
				m.newMetadata.DataType = "cards"
				return m, cmd
				/* default:
				m.currentStage = "MainMenu"
				m.errorMessage = "wrong selected type"
				return m, cmd*/
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
				return m, cmd
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
				m.currentStage = "WriteToServer"
				m.loginAndPasswordData.Password = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}

	case "WriteNumber":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteCardholderName"
				m.cardData.CardNumber = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "WriteCardholderName":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteExpirationDate"
				m.cardData.CardholderName = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "WriteExpirationDate":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteCode"
				m.cardData.ExpiredAt = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "WriteCode":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "WriteToServer"
				m.cardData.Code = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "WriteToServer":
		msg := handleWrite(m)
		m.currentStage = msg.NextStageNameKey
		m.errorMessage = msg.ErrorMessage
		return m, cmd

	case "Edit":
		m.targetObjectMetadata, m.targetObjectIndex = getMetadataByName(m)
		if m.targetObjectIndex < 0 {
			m.errorMessage = "no such name"
			m.currentStage = "MainMenu"
		}
		m.currentStage = "EditDescription"
		return m, nil
	case "EditDescription":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				switch m.targetObjectMetadata.DataType {
				case "passwords":
					m.currentStage = "EditLogin"
					m.newMetadata.Description = m.textInput.Value()
					m.textInput.SetValue("")
				case "cards":
					m.currentStage = "EditNumber"
					m.newMetadata.Description = m.textInput.Value()
					m.textInput.SetValue("")
				}
				return m, cmd
			}
		}
		return m, nil
	case "EditLogin":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditPassword"
				m.loginAndPasswordData.Login = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "EditPassword":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditToServer"
				m.loginAndPasswordData.Password = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}

	case "EditNumber":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditCardholderName"
				m.cardData.CardNumber = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "EditCardholderName":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditExpirationDate"
				m.cardData.CardholderName = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "EditExpirationDate":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditCode"
				m.cardData.ExpiredAt = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "EditCode":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.currentStage = "EditToServer"
				m.cardData.Code = m.textInput.Value()
				m.textInput.SetValue("")
				return m, cmd
			}
		}
	case "EditToServer":
		msg := handleEdit(m)
		m.currentStage = msg.NextStageNameKey
		m.errorMessage = msg.ErrorMessage
		return m, cmd
	case "DataSaved":
		m.currentStage = "MainMenu"
		return m, cmd
	case "List":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.currentStage = "MainMenu"
				return m, cmd
			}
		}
	case "Read":
		msg := readHandle(m)
		m.currentStage = msg.NextStageNameKey
		m.errorMessage = msg.ErrorMessage
		return m, cmd
	case "ReadComplete":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.currentStage = "MainMenu"
				return m, cmd
			}
		}
	case "Delete":
		msg := deleteHandle(m)
		m.currentStage = msg.NextStageNameKey
		m.errorMessage = msg.ErrorMessage
		return m, cmd
	case "DeleteComplete":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.currentStage = "MainMenu"
				return m, cmd
			}
		}
	}

	return m, tick()
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
		m.textInput.Placeholder = "Name"
		return fmt.Sprintf(
			"Input name of the data: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteDescription":
		m.textInput.Placeholder = "Description"
		return fmt.Sprintf(
			"Input description of the data: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "EditDescription":
		m.textInput.Placeholder = "Description"
		return fmt.Sprintf(
			"Input new description of the data: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "SelectDataType":
		s = "press '1' to add login and password data or press '2' to add card data"
	case "WriteLogin", "EditLogin":
		return fmt.Sprintf(
			"Input login: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WritePassword", "EditPassword":
		m.textInput.Placeholder = "Password"
		m.textInput.EchoMode = textinput.EchoPassword
		m.textInput.EchoCharacter = '*'
		return fmt.Sprintf(
			"Input password: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteNumber", "EditNumber":
		return fmt.Sprintf(
			"Input number: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteCardholderName", "EditCardholderName":
		return fmt.Sprintf(
			"Input cardholder name: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteExpirationDate", "EditExpirationDate":
		return fmt.Sprintf(
			"Input expiration date: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "WriteCode", "EditCode":
		return fmt.Sprintf(
			"Input code: \n\n%s\n\n",
			m.textInput.View(),
		) + "\n"
	case "EditToServer":
		s = "writing to server"
	case "DataSaved":
		s = "Data added"
	case "List":
		s = drawList(m)
	case "Read":
		s = "Reading"
	case "ReadComplete":
		s = fmt.Sprintf(
			"Your data: \n\n%s\n\n ",
			*m.outputData,
		) + "\n"
	case "Delete":
		s = "Deletion"
	case "DeleteComplete":
		s = "Data deleted"
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

func handleLogin(m model) stageCompleteMsg {
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
		msg.NextStageNameKey = "Sync"
		return msg
	}
	msg.ErrorMessage = "server error, unexpected status: " + fmt.Sprint(status)
	msg.NextStageNameKey = "AuthFailed"
	return msg
}

func handleRegister(m model) stageCompleteMsg {
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

func handleSync(m model) stageCompleteMsg {
	status, userMetadata, err := m.clientEnv.SyncHandle()
	*m.userMetadata = userMetadata

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
			*m.targetObjectName = commandSlice[1]
		default:
			msg.ErrorMessage = "Unknown command"
			msg.NextStageNameKey = "MainMenu"
		}
		return msg
	}
}

func handleWrite(m model) stageCompleteMsg {
	var msg stageCompleteMsg

	var data []byte

	switch m.newMetadata.DataType {
	case "passwords":
		bytes, err := json.Marshal(m.loginAndPasswordData)
		m.loginAndPasswordData = &gophmodel.LoginAndPasswordData{}
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "MainMenu"
		}
		data = bytes

	case "cards":
		bytes, err := json.Marshal(m.cardData)
		m.cardData = &gophmodel.CardData{}
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "MainMenu"
		}
		data = bytes
	}

	status, metadata, err := m.clientEnv.WriteHandle(m.newMetadata, data)
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	if status == 200 {
		msg.ErrorMessage = ""
		msg.NextStageNameKey = "DataSaved"
		*m.userMetadata = append(*m.userMetadata, metadata)
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "MainMenu"
		return msg
	}

	return msg
}

func getMetadataByName(m model) (gophmodel.Metadata, int) {
	var metadataToEdit gophmodel.Metadata

	metadataIndex := -1

	for i, metadata := range *m.userMetadata {
		if metadata.Name == *m.targetObjectName {
			metadataToEdit = metadata
			metadataIndex = i
		}
	}
	return metadataToEdit, metadataIndex
}

func nameAlreadyExists(m model, name string) bool {
	for _, metadata := range *m.userMetadata {
		if metadata.Name == name {
			return true
		}
	}
	return false
}

func handleEdit(m model) stageCompleteMsg {
	var msg stageCompleteMsg

	metadataToEdit := m.targetObjectMetadata

	var data []byte

	switch metadataToEdit.DataType {
	case "passwords":
		bytes, err := json.Marshal(m.loginAndPasswordData)
		m.loginAndPasswordData = &gophmodel.LoginAndPasswordData{}
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "MainMenu"
		}
		data = bytes
	case "cards":
		bytes, err := json.Marshal(m.cardData)
		m.cardData = &gophmodel.CardData{}
		if err != nil {
			msg.ErrorMessage = err.Error()
			msg.NextStageNameKey = "MainMenu"
		}
		data = bytes
	}

	status, metadata, err := m.clientEnv.EditHandle(metadataToEdit, m.newMetadata, data)
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	if status == 200 {
		msg.ErrorMessage = ""
		msg.NextStageNameKey = "DataSaved"
		(*m.userMetadata)[m.targetObjectIndex] = metadata
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "MainMenu"
		return msg
	}

	return msg
}

func drawList(m model) string {
	var sb strings.Builder
	sb.WriteString("Name, Description, Data Type, Changed, Created\n\n")
	info := *m.userMetadata
	count := len(info)
	for i := 0; i < count; i++ {
		sb.WriteString(fmt.Sprintf("Name: %s , Description: %s , Data Type: %s , Changed: %s , Created: %s\n\n",
			info[i].Name,
			info[i].Description,
			info[i].DataType,
			info[i].Changed,
			info[i].Created,
		))
	}
	return sb.String()
}

func readHandle(m model) stageCompleteMsg {
	var msg stageCompleteMsg

	var metadataToRead gophmodel.Metadata

	for _, metadata := range *m.userMetadata {
		if metadata.Name == *m.targetObjectName {
			metadataToRead = metadata
		}
	}

	if metadataToRead == (gophmodel.Metadata{}) {
		msg.ErrorMessage = "no such name"
		msg.NextStageNameKey = "MainMenu"
	}

	status, data, err := m.clientEnv.ReadHandle(metadataToRead)
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	msg.ErrorMessage = ""
	msg.NextStageNameKey = "ReadComplete"
	*m.outputData = string(data)

	return msg
}

func readFileHandle(m model) stageCompleteMsg {
	var msg stageCompleteMsg

	var metadataToRead gophmodel.Metadata

	for _, metadata := range *m.userMetadata {
		if metadata.Name == *m.targetObjectName {
			metadataToRead = metadata
		}
	}

	if metadataToRead == (gophmodel.Metadata{}) {
		msg.ErrorMessage = "no such name"
		msg.NextStageNameKey = "MainMenu"
	}

	status, data, err := m.clientEnv.ReadHandle(metadataToRead)
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	msg.ErrorMessage = ""
	msg.NextStageNameKey = "ReadComplete"
	
	*m.outputData = string(data)

	return msg
}

func deleteHandle(m model) stageCompleteMsg {
	var msg stageCompleteMsg

	var metadataToDelete gophmodel.Metadata

	deleteIndex := -1

	for i, metadata := range *m.userMetadata {
		if metadata.Name == *m.targetObjectName {
			metadataToDelete = metadata
			deleteIndex = i
		}
	}

	if deleteIndex == -1 || metadataToDelete == (gophmodel.Metadata{}) {
		msg.ErrorMessage = "no such name"
		msg.NextStageNameKey = "MainMenu"
	}

	status, err := m.clientEnv.DeleteHandle(metadataToDelete)
	if err != nil {
		msg.ErrorMessage = err.Error()
		msg.NextStageNameKey = "MainMenu"
		return msg
	}
	if status != 200 {
		msg.ErrorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		msg.NextStageNameKey = "MainMenu"
		return msg
	}

	if status == 200 {
		msg.ErrorMessage = ""
		msg.NextStageNameKey = "DeleteComplete"
		(*m.userMetadata) = append((*m.userMetadata)[:deleteIndex], (*m.userMetadata)[deleteIndex+1:]...)
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

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
