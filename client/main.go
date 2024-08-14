package main

import (
	"encoding/json"
	"fmt"
	handler "gophkeep/client/internal/handler"
	gophmodel "gophkeep/internal/model"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	UI           ui
	ClientEnv    *handler.ClientEnv
	TargetObject *targetObject
	NewData      *newData

	stageState *stageState

	OutputData *string

	UserMetadata *[]gophmodel.Metadata
	TextInput    textinput.Model
}

type stageState struct {
	nextStage    string
	errorMessage string
}

type ui struct {
	Width  int
	Height int
}

type targetObject struct {
	Metadata gophmodel.Metadata
	Index    int
	Name     string
}

type newData struct {
	LoginInfo            gophmodel.SimpleAccountData
	AuthType             string
	Metadata             gophmodel.SimpleMetadata
	LoginAndPasswordData gophmodel.LoginAndPasswordData
	CardData             gophmodel.CardData
	FilePath             string
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 255
	outputData := ""

	return model{
		TextInput:    ti,
		stageState:   &stageState{},
		ClientEnv:    &handler.ClientEnv{},
		UserMetadata: &[]gophmodel.Metadata{},

		TargetObject: &targetObject{},
		OutputData:   &outputData,

		NewData: &newData{},
	}
}

func (m model) Init() tea.Cmd {
	return startAppCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UI.Width = msg.Width
		m.UI.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case startAppMsg:
		{
			m.handlePingServer()
			return m, cmd
		}
	}

	switch m.stageState.nextStage {
	case "PingFail":
		return m.updatePingFail(msg, cmd)
	case "SignInChoise":
		return m.updateSignInChoise(msg, cmd)
	case "LoginRegisterInputs":
		return m.updateLoginRegisterInputs(msg, cmd)
	case "PasswordInput":
		return m.updatePasswordInput(msg, cmd)
	case "Auth":
		m.updateAuth(cmd)
		return m, cmd
	case "AuthFailed":
		return m.updateAuthFailed(cmd)
	case "Sync":
		m.updateSync(cmd)
		return m, cmd
	case "MainMenu":
		return m.updateMainMenu(msg, cmd)
	case "Write":
		return m.updateWrite()
	case "WriteName":
		return m.updateWriteName(msg, cmd)
	case "WriteDescription":
		return m.updateWriteDescription(msg, cmd)
	case "SelectDataType":
		return m.updateSelectDataType(msg, cmd)
	case "WriteFile":
		return m.updateWriteFile(msg, cmd)
	case "EditFile":
		return m.updateEditFile(msg, cmd)
	case "WriteLogin":
		return m.updateWriteLogin(msg, cmd)
	case "WritePassword":
		return m.updateWritePassword(msg, cmd)
	case "WriteNumber":
		return m.updateWriteNumber(msg, cmd)
	case "WriteCardholderName":
		return m.updateWriteCardholderName(msg, cmd)
	case "WriteExpirationDate":
		return m.updateWriteExpirationDate(msg, cmd)
	case "WriteCode":
		return m.updateWriteCode(msg, cmd)
	case "WriteFileToServer":
		m.updateWriteFileToServer(cmd)
		return m, cmd
	case "EditFileToServer":
		m.updateEditFileToServer(cmd)
		return m, cmd
	case "WriteToServer":
		m.updateWriteToServer(cmd)
		return m, cmd
	case "Edit":
		return m.updateEdit()
	case "EditDescription":
		return m.updateEditDescription(msg, cmd)
	case "EditLogin":
		return m.updateEditLogin(msg, cmd)
	case "EditPassword":
		return m.updateEditPassword(msg, cmd)
	case "EditNumber":
		return m.updateEditNumber(msg, cmd)
	case "EditCardholderName":
		return m.updateEditCardholderName(msg, cmd)
	case "EditExpirationDate":
		return m.updateEditExpirationDate(msg, cmd)
	case "EditCode":
		return m.updateEditCode(msg, cmd)
	case "EditToServer":
		m.updateEditToServer(cmd)
		return m, cmd
	case "DataSaved":
		return m.updateDataSaved(cmd)
	case "List":
		return m.updateList(msg, cmd)
	case "Read":
		m.updateRead(cmd)
		return m, cmd
	case "ReadComplete":
		return m.updateReadComplete(msg, cmd)
	case "ReadFileComplete":
		return m.updateReadFileComplete(msg, cmd)
	case "Delete":
		m.updateDelete(cmd)
		return m, cmd
	case "DeleteComplete":
		return m.updateDeleteComplete(msg, cmd)
	}

	return m, cmd
}

func (m model) updateDeleteComplete(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "MainMenu"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateDelete(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.deleteHandle()
	return m, cmd
}

func (m model) updateReadFileComplete(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "MainMenu"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateReadComplete(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "MainMenu"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateRead(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.readHandle()
	return m, cmd
}

func (m model) updateList(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "MainMenu"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateDataSaved(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.stageState.nextStage = "MainMenu"
	return m, cmd
}

func (m model) updateEditToServer(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.handleEdit()
	return m, cmd
}

func (m model) updateEditCode(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditToServer"
			m.NewData.CardData.Code = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditExpirationDate(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditCode"
			m.NewData.CardData.ExpiredAt = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditCardholderName(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditExpirationDate"
			m.NewData.CardData.CardholderName = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditNumber(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditCardholderName"
			m.NewData.CardData.CardNumber = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditPassword(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditToServer"
			m.NewData.LoginAndPasswordData.Password = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditLogin(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditPassword"
			m.NewData.LoginAndPasswordData.Login = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditDescription(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			switch m.TargetObject.Metadata.DataType {
			case "passwords":
				m.stageState.nextStage = "EditLogin"
				m.NewData.Metadata.Description = m.TextInput.Value()
				m.TextInput.SetValue("")
			case "cards":
				m.stageState.nextStage = "EditNumber"
				m.NewData.Metadata.Description = m.TextInput.Value()
				m.TextInput.SetValue("")
			case "files":
				m.stageState.nextStage = "EditFile"
				m.NewData.Metadata.Description = m.TextInput.Value()
				m.TextInput.SetValue("")
			}
			return m, cmd
		}
	}
	return m, nil
}

func (m model) updateEdit() (tea.Model, tea.Cmd) {
	m.TargetObject.Metadata, m.TargetObject.Index = getMetadataByName(m)
	if m.TargetObject.Index < 0 {
		m.stageState.errorMessage = "no such name"
		m.stageState.nextStage = "MainMenu"
	}
	m.stageState.nextStage = "EditDescription"
	return m, nil
}

func (m model) updateWriteToServer(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.handleWrite()
	return m, cmd
}

func (m model) updateEditFileToServer(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.handleEditFile()
	return m, cmd
}

func (m model) updateWriteFileToServer(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.handleWriteFile()
	return m, cmd
}

func (m model) updateWriteCode(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteToServer"
			m.NewData.CardData.Code = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateWriteExpirationDate(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteCode"
			m.NewData.CardData.ExpiredAt = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}

	return m, cmd
}

func (m model) updateWriteCardholderName(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteExpirationDate"
			m.NewData.CardData.CardholderName = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}

	return m, cmd
}

func (m model) updateWriteNumber(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteCardholderName"
			m.NewData.CardData.CardNumber = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateWritePassword(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteToServer"
			m.NewData.LoginAndPasswordData.Password = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateWriteLogin(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WritePassword"
			m.NewData.LoginAndPasswordData.Login = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateEditFile(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "EditFileToServer"
			m.NewData.FilePath = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateWriteFile(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "WriteFileToServer"
			m.NewData.FilePath = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateSelectDataType(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.stageState.nextStage = "WriteLogin"
			m.NewData.Metadata.DataType = "passwords"
			return m, cmd
		case "2":
			m.stageState.nextStage = "WriteNumber"
			m.NewData.Metadata.DataType = "cards"
			return m, cmd
		case "3":
			m.stageState.nextStage = "WriteFile"
			m.NewData.Metadata.DataType = "files"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateWriteDescription(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "SelectDataType"
			m.NewData.Metadata.Description = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, nil
}

func (m model) updateWriteName(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.NewData.Metadata.Name = m.TextInput.Value()
			if nameAlreadyExists(m, m.NewData.Metadata.Name) {
				m.stageState.errorMessage = "you already use that data name"
				m.stageState.nextStage = "MainMenu"
			} else {
				m.stageState.nextStage = "WriteDescription"
			}
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, nil
}

func (m model) updateWrite() (tea.Model, tea.Cmd) {
	m.stageState.nextStage = "WriteName"
	return m, nil
}

func (m *model) updateMainMenu(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.handleMainMenuCommand()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateSync(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.handleSync()
	return m, cmd
}

func (m model) updateAuthFailed(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.stageState.nextStage = "SignInChoise"
	return m, cmd
}

func (m model) updateAuth(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.NewData.AuthType == "login" {
		m.handleLogin()
		return m, cmd
	}

	if m.NewData.AuthType == "register" {
		m.handleRegister()
		return m, cmd
	}
	return m, cmd
}

func (m model) updatePasswordInput(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "Auth"
			m.NewData.LoginInfo.Password = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateLoginRegisterInputs(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			m.stageState.nextStage = "PasswordInput"
			m.NewData.LoginInfo.Login = m.TextInput.Value()
			m.TextInput.SetValue("")
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updateSignInChoise(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			m.stageState.nextStage = "LoginRegisterInputs"
			m.NewData.AuthType = "login"
			m.TextInput.Placeholder = "Enter your login here"
			return m, cmd
		case "r":
			m.stageState.nextStage = "LoginRegisterInputs"
			m.NewData.AuthType = "register"
			m.TextInput.Placeholder = "Enter your new login here"
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) updatePingFail(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.handlePingServer()
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) View() string {
	s := m.stageState.nextStage + " stage"
	switch m.stageState.nextStage {
	case "PingServer":
		s = "Connecting to server"
	case "SignInChoise":
		s = "type 'l' or 'r' to login or register"
		if len(m.stageState.errorMessage) != 0 {
			s = m.stageState.errorMessage + "\n" + s
		}
	case "PingFail":
		s = m.stageState.errorMessage + "\n Could not connect to the server" +
			"\n Press Enter to retry"

	case "LoginRegisterInputs":
		return fmt.Sprintf(
			"Input your login:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "PasswordInput":
		m.TextInput.Placeholder = "Password"
		m.TextInput.EchoMode = textinput.EchoPassword
		m.TextInput.EchoCharacter = '*'
		return fmt.Sprintf(
			"Input your password: \n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "Sync":
		s = "Loading data from server"
	case "SyncFail":
		s = "Sync failed " + m.stageState.errorMessage
	case "MainMenu":
		errorMessage := ""
		if len(m.stageState.errorMessage) != 0 {
			errorMessage = m.stageState.errorMessage + "\n\n"
		}
		m.TextInput.Placeholder = "Type command here"
		s = errorMessage + "Menu\n\n" + "type:\n\nread <name> to read your saved data" +
			"\n\nwrite to add new data" +
			"\n\nlist to view all names and descriptions of your data" +
			"\n\ndelete <name> to delete data" +
			"\n\nedit <name> to edit data \n\n" + m.TextInput.View()
	case "WriteName":
		m.TextInput.Placeholder = "Name"
		return fmt.Sprintf(
			"Input name of the data:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteDescription":
		m.TextInput.Placeholder = "Description"
		return fmt.Sprintf(
			"Input description of the data:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "EditDescription":
		m.TextInput.Placeholder = "Description"
		return fmt.Sprintf(
			"Input new description of the data:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "SelectDataType":
		s = "press '1' to add login and password data or press '2' to add card data or '3' to add file"
	case "WriteLogin", "EditLogin":
		return fmt.Sprintf(
			"Input login:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WritePassword", "EditPassword":
		m.TextInput.Placeholder = "Password"
		m.TextInput.EchoMode = textinput.EchoPassword
		m.TextInput.EchoCharacter = '*'
		return fmt.Sprintf(
			"Input password:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteNumber", "EditNumber":
		return fmt.Sprintf(
			"Input number: \n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteCardholderName", "EditCardholderName":
		return fmt.Sprintf(
			"Input cardholder name:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteExpirationDate", "EditExpirationDate":
		return fmt.Sprintf(
			"Input expiration date:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteCode", "EditCode":
		return fmt.Sprintf(
			"Input code:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "WriteFile":
		m.TextInput.Placeholder = "File path"
		return fmt.Sprintf(
			"Input file path:\n\n%s\n\n",
			m.TextInput.View(),
		) + "\n"
	case "EditToServer":
		s = "writing to server"
	case "DataSaved":
		s = "Data added"
	case "List":
		s = m.drawList()
	case "Read":
		s = "Reading"
	case "ReadComplete":
		s = fmt.Sprintf(
			"Your data:\n\n%s\n\n",
			*m.OutputData,
		) + "\n"
	case "ReadFileComplete":
		s = fmt.Sprintf(
			"Path to file:\n\n%s\n\n",
			*m.OutputData,
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

func (m model) handleLogin() {
	status, err := m.ClientEnv.HandleLogin(m.NewData.LoginInfo)
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "AuthFailed"
		return
	}
	if status == http.StatusUnauthorized {
		m.stageState.errorMessage = "no such login and password pair found"
		m.stageState.nextStage = "AuthFailed"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "Sync"
		return
	}
	m.stageState.errorMessage = "server error, unexpected status: " + fmt.Sprint(status)
	m.stageState.nextStage = "AuthFailed"
}

func (m model) handleRegister() {
	status, err := m.ClientEnv.HandleRegister(m.NewData.LoginInfo)
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "AuthFailed"
		return
	}
	if status == http.StatusConflict {
		m.stageState.errorMessage = "login alredy in use"
		m.stageState.nextStage = "AuthFailed"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "AuthSuccess"
		return
	}
	m.stageState.errorMessage = "server error, unexpected status: " + fmt.Sprint(status)
	m.stageState.nextStage = "AuthFailed"
}

func (m model) handlePingServer() {
	status, err := m.ClientEnv.HandlePingServer()
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "PingFail"
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "status is not OK, it is: " + fmt.Sprint(status)
		m.stageState.nextStage = "PingFail"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "SignInChoise"
		return
	}
}

func (m model) handleSync() {
	status, userMetadata, err := m.ClientEnv.HandleSync()
	*m.UserMetadata = userMetadata

	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "SyncFail"
		return
	}
	if status == http.StatusNoContent || status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "MainMenu"
		return
	}
}

func (m model) handleMainMenuCommand() {
	command := m.TextInput.Value()
	commandSlice := strings.Fields(command)
	switch len(commandSlice) {
	case 0:
		m.stageState.errorMessage = "command didn't have any words"
		m.stageState.nextStage = "MainMenu"
	case 1:
		switch commandSlice[0] {
		case "write":
			m.stageState.errorMessage = ""
			m.stageState.nextStage = "Write"
		case "list":
			m.stageState.errorMessage = ""
			m.stageState.nextStage = "List"
		default:
			m.stageState.errorMessage = "Unknown command"
			m.stageState.nextStage = "MainMenu"
		}
	case 2:
		switch commandSlice[0] {
		case "read":
			m.stageState.errorMessage = ""
			m.stageState.nextStage = "Read"
		case "delete":
			m.stageState.errorMessage = ""
			m.stageState.nextStage = "Delete"
		case "edit":
			m.stageState.errorMessage = ""
			m.stageState.nextStage = "Edit"
		default:
			m.stageState.errorMessage = "Unknown command"
			m.stageState.nextStage = "MainMenu"
		}
		m.TargetObject.Name = commandSlice[1]
	default:
		m.stageState.errorMessage = "Unknown command"
		m.stageState.nextStage = "MainMenu"
	}
}

func (m model) handleWrite() {
	var data []byte

	switch m.NewData.Metadata.DataType {
	case "passwords":
		bytes, err := json.Marshal(m.NewData.LoginAndPasswordData)
		m.NewData.LoginAndPasswordData = gophmodel.LoginAndPasswordData{}
		if err != nil {
			m.stageState.errorMessage = err.Error()
			m.stageState.nextStage = "MainMenu"
		}
		data = bytes

	case "cards":
		bytes, err := json.Marshal(m.NewData.CardData)
		m.NewData.CardData = gophmodel.CardData{}
		if err != nil {
			m.stageState.errorMessage = err.Error()
			m.stageState.nextStage = "MainMenu"
		}
		data = bytes
	}

	status, metadata, err := m.ClientEnv.HandleWrite(m.NewData.Metadata, data)
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "MainMenu"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "DataSaved"
		*m.UserMetadata = append(*m.UserMetadata, metadata)
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		m.stageState.nextStage = "MainMenu"
		return
	}
}

func (m model) handleWriteFile() {
	status, metadata, err := m.ClientEnv.HandleWriteFile(m.NewData.Metadata, []byte(m.NewData.FilePath))

	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "MainMenu"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "DataSaved"
		*m.UserMetadata = append(*m.UserMetadata, metadata)
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		m.stageState.nextStage = "MainMenu"
		return
	}
}

func (m model) handleEditFile() {
	metadataToEdit := m.TargetObject.Metadata

	status, metadata, err := m.ClientEnv.HandleEditFile(metadataToEdit, m.NewData.Metadata, []byte(m.NewData.FilePath))
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "MainMenu"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "DataSaved"
		(*m.UserMetadata)[m.TargetObject.Index] = metadata
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		m.stageState.nextStage = "MainMenu"
		return
	}
}

func getMetadataByName(m model) (gophmodel.Metadata, int) {
	var metadataToEdit gophmodel.Metadata

	metadataIndex := -1

	for i, metadata := range *m.UserMetadata {
		if metadata.Name == m.TargetObject.Name {
			metadataToEdit = metadata
			metadataIndex = i
		}
	}
	return metadataToEdit, metadataIndex
}

func nameAlreadyExists(m model, name string) bool {
	for _, metadata := range *m.UserMetadata {
		if metadata.Name == name {
			return true
		}
	}
	return false
}

func (m model) handleEdit() {
	metadataToEdit := m.TargetObject.Metadata

	var data []byte

	switch metadataToEdit.DataType {
	case "passwords":
		bytes, err := json.Marshal(m.NewData.LoginAndPasswordData)
		m.NewData.LoginAndPasswordData = gophmodel.LoginAndPasswordData{}
		if err != nil {
			m.stageState.errorMessage = err.Error()
			m.stageState.nextStage = "MainMenu"
		}
		data = bytes
	case "cards":
		bytes, err := json.Marshal(m.NewData.CardData)
		m.NewData.CardData = gophmodel.CardData{}
		if err != nil {
			m.stageState.errorMessage = err.Error()
			m.stageState.nextStage = "MainMenu"
		}
		data = bytes
	}

	status, metadata, err := m.ClientEnv.HandleEdit(metadataToEdit, m.NewData.Metadata, data)
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "MainMenu"
		return
	}
	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "DataSaved"
		(*m.UserMetadata)[m.TargetObject.Index] = metadata
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		m.stageState.nextStage = "MainMenu"
		return
	}
}

func (m model) drawList() string {
	var sb strings.Builder
	sb.WriteString("Name, Description, Data Type, Changed, Created\n\n")
	info := *m.UserMetadata
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

func (m model) readHandle() {
	var metadataToRead gophmodel.Metadata

	for _, metadata := range *m.UserMetadata {
		if metadata.Name == m.TargetObject.Name {
			metadataToRead = metadata
		}
	}

	if metadataToRead == (gophmodel.Metadata{}) {
		m.stageState.errorMessage = "no such name"
		m.stageState.nextStage = "MainMenu"
	}

	if metadataToRead.DataType == "files" {
		status, filePath, err := m.ClientEnv.HandleReadFile(metadataToRead)
		if err != nil {
			m.stageState.errorMessage = "Could not request file " + err.Error()
			m.stageState.nextStage = "MainMenu"
			return
		}
		if status != http.StatusOK {
			m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
			m.stageState.nextStage = "MainMenu"
			return
		}
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "ReadFileComplete"

		*m.OutputData = string(filePath)

		return
	} else {
		status, data, err := m.ClientEnv.HandleRead(metadataToRead)
		if err != nil {
			m.stageState.errorMessage = "Could not request data: " + metadataToRead.StaticID + " " + err.Error()
			m.stageState.nextStage = "MainMenu"
			return
		}
		if status != http.StatusOK {
			m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
			m.stageState.nextStage = "MainMenu"
			return
		}
		if metadataToRead.DataType == "cards" {
			var cardData gophmodel.CardData

			if err = json.Unmarshal(data, &cardData); err != nil {
				m.stageState.errorMessage = "could not unmarshal JSON"
				m.stageState.nextStage = "MainMenu"
			}

			s := fmt.Sprintf("Number: %s Card holder: %s\n\nExpiry date: %s Code: %s", cardData.CardNumber, cardData.CardholderName, cardData.ExpiredAt, cardData.Code)

			*m.OutputData = string(s)
		} else if metadataToRead.DataType == "passwords" {
			var password gophmodel.LoginAndPasswordData

			if err = json.Unmarshal(data, &password); err != nil {
				m.stageState.errorMessage = "could not unmarshal JSON"
				m.stageState.nextStage = "MainMenu"
			}

			s := fmt.Sprintf("Login: %s\nPassword: %s", password.Login, password.Password)

			*m.OutputData = string(s)
		} else {
			m.stageState.errorMessage = "Could not find data of this datatype: " + metadataToRead.DataType
			m.stageState.nextStage = "MainMenu"
		}

		m.stageState.errorMessage = ""
		m.stageState.nextStage = "ReadComplete"

		return
	}
}

func (m model) deleteHandle() {
	var metadataToDelete gophmodel.Metadata

	deleteIndex := -1

	for i, metadata := range *m.UserMetadata {
		if metadata.Name == m.TargetObject.Name {
			metadataToDelete = metadata
			deleteIndex = i
		}
	}

	if deleteIndex == -1 || metadataToDelete == (gophmodel.Metadata{}) {
		m.stageState.errorMessage = "no such name"
		m.stageState.nextStage = "MainMenu"
	}

	status, err := m.ClientEnv.HandleDelete(metadataToDelete)
	if err != nil {
		m.stageState.errorMessage = err.Error()
		m.stageState.nextStage = "MainMenu"
		return
	}
	if status != http.StatusOK {
		m.stageState.errorMessage = "Something went wrong with status: " + fmt.Sprint(status)
		m.stageState.nextStage = "MainMenu"
		return
	}

	if status == http.StatusOK {
		m.stageState.errorMessage = ""
		m.stageState.nextStage = "DeleteComplete"
		(*m.UserMetadata) = append((*m.UserMetadata)[:deleteIndex], (*m.UserMetadata)[deleteIndex+1:]...)
		return
	}
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
