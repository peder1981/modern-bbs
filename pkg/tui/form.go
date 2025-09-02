package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"modern-bbs/internal/database"
)

// formInput define a interface para campos de formulário.
// formInput define a interface para campos de formulário.
type formInput interface {
	Focus() tea.Cmd
	Blur()
	View() string
	Value() string
	Update(tea.Msg) (formInput, tea.Cmd)
}

// FormField encapsula um campo de formulário e seu nome.
type FormField struct {
	Name  string
	Input formInput
}

// formModel é um modelo genérico para formulários de criação de conteúdo.
type formModel struct {
	parent       *mainModel
	title        string
	fields       []FormField
	focusIndex   int
	submitAction func(map[string]string) tea.Cmd
	quitting     bool
}

func newTextInput(placeholder string) formInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 150
	ti.Width = 80
	return &TextInput{Model: ti}
}

func newTextArea(placeholder string) formInput {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(5)
	return &TextArea{Model: ta}
}

// NewTopicFormModel cria um formulário para um novo tópico.
func NewTopicFormModel(parent *mainModel, forum *database.Forum) *formModel {
	titleInput := newTextInput("Título do Tópico")
	titleInput.Focus()

	fields := []FormField{
		{Name: "Título", Input: titleInput},
	}

	return &formModel{
		parent:     parent,
		title:      fmt.Sprintf("Novo Tópico em '%s'", forum.Name),
		fields:     fields,
		focusIndex: 0,
		submitAction: func(values map[string]string) tea.Cmd {
			topicTitle := values["Título"]
			if topicTitle == "" {
				return func() tea.Msg { return statusMessage{success: false, message: "O título não pode estar vazio."} }
			}
			user, _, err := database.GetUserByUsername(parent.User)
			if err != nil {
				return func() tea.Msg { return errorMsg{err} }
			}
			err = database.CreateTopic(int(forum.ID), int(user.ID), topicTitle)
			if err != nil {
				return func() tea.Msg { return errorMsg{err} }
			}
			return func() tea.Msg { return topicCreatedMsg{forum: forum} }
		},
	}
}

// NewPostFormModel cria um formulário para um novo post (resposta).
func NewPostFormModel(parent *mainModel, topic *database.Topic) *formModel {
	postTextArea := newTextArea("Escreva sua resposta...")
	postTextArea.Focus()

	fields := []FormField{
		{Name: "Conteúdo", Input: postTextArea},
	}

	return &formModel{
		parent:     parent,
		title:      fmt.Sprintf("Re: %s", topic.Title),
		fields:     fields,
		focusIndex: 0,
		submitAction: func(values map[string]string) tea.Cmd {
			postContent := values["Conteúdo"]
			if postContent == "" {
				return func() tea.Msg { return statusMessage{success: false, message: "O conteúdo não pode estar vazio."} }
			}
			user, _, err := database.GetUserByUsername(parent.User)
			if err != nil {
				return func() tea.Msg { return errorMsg{err} }
			}
			err = database.CreatePost(topic.ID, int(user.ID), postContent)
			if err != nil {
				return func() tea.Msg { return errorMsg{err} }
			}
			return func() tea.Msg { return postCreatedMsg{topic: topic} }
		},
	}
}

// NewUserFormModel cria um formulário para um novo usuário.
func NewUserFormModel(parent *mainModel) *formModel {
	usernameInput := newTextInput("Nome de Usuário")
	passwordInput := newTextInput("Senha")
	roleInput := newTextInput("Papel (user, moderator, admin)")

	passwordInput.(*TextInput).EchoMode = textinput.EchoPassword
	usernameInput.Focus()

	fields := []FormField{
		{Name: "Username", Input: usernameInput},
		{Name: "Password", Input: passwordInput},
		{Name: "Role", Input: roleInput},
	}

	return &formModel{
		parent:     parent,
		title:      "Criar Novo Usuário",
		fields:     fields,
		focusIndex: 0,
		submitAction: func(values map[string]string) tea.Cmd {
			username := values["Username"]
			password := values["Password"]
			role := values["Role"]

			return func() tea.Msg {
				if username == "" || password == "" || role == "" {
					return statusMessage{success: false, message: "Todos os campos são obrigatórios."}
				}

				if role != "user" && role != "moderator" && role != "admin" {
					return statusMessage{success: false, message: "Papel inválido. Use 'user', 'moderator' ou 'admin'."}
				}

				_, err := database.CreateUser(username, password)
				if err != nil {
					return errorMsg{err}
				}

				err = database.SetUserRole(username, role)
				if err != nil {
					return errorMsg{fmt.Errorf("usuário criado, mas falha ao definir o papel: %w", err)}
				}

				return statusMessage{success: true, message: fmt.Sprintf("Usuário '%s' criado com sucesso!", username)}
			}
		},
	}
}

func NewChangePasswordFormModel(parent *mainModel) *formModel {
	currentPasswordInput := newTextInput("Senha Atual")
	newPasswordInput := newTextInput("Nova Senha")

	currentPasswordInput.(*TextInput).EchoMode = textinput.EchoPassword
	newPasswordInput.(*TextInput).EchoMode = textinput.EchoPassword
	currentPasswordInput.Focus()

	fields := []FormField{
		{Name: "CurrentPassword", Input: currentPasswordInput},
		{Name: "NewPassword", Input: newPasswordInput},
	}

	return &formModel{
		parent:     parent,
		title:      "Alterar Senha",
		fields:     fields,
		focusIndex: 0,
		submitAction: func(values map[string]string) tea.Cmd {
			currentPassword := values["CurrentPassword"]
			newPassword := values["NewPassword"]
			return func() tea.Msg {
				if currentPassword == "" || newPassword == "" {
					return statusMessage{success: false, message: "Os campos de senha não podem estar vazios."}
				}
				err := database.UpdateUserPassword(parent.User, currentPassword, newPassword)
				if err != nil {
					return errorMsg{err}
				}
				return passwordUpdatedMsg{}
			}
		},
	}
}

// NewForumFormModel cria um formulário para um novo fórum.
// NewEditForumFormModel cria um formulário para editar um fórum existente.
func NewEditForumFormModel(parent *mainModel, forum *database.Forum) *formModel {
	nameInput := newTextInput("Nome do Fórum")
	nameInput.(*TextInput).SetValue(forum.Name)

	descArea := newTextArea("Descrição do Fórum")
	descArea.(*TextArea).SetValue(forum.Description)

	nameInput.Focus()

	fields := []FormField{
		{Name: "Nome", Input: nameInput},
		{Name: "Descrição", Input: descArea},
	}

	return &formModel{
		parent:     parent,
		title:      fmt.Sprintf("Editando Fórum: %s", forum.Name),
		fields:     fields,
		focusIndex: 0,
		submitAction: func(values map[string]string) tea.Cmd {
			return func() tea.Msg {
				err := database.UpdateForum(forum.ID, values["Nome"], values["Descrição"])
				if err != nil {
					return statusMessage{success: false, message: "Erro ao atualizar fórum: " + err.Error()}
				}
				return statusMessage{success: true, message: fmt.Sprintf("Fórum '%s' atualizado com sucesso!", values["Nome"])}
			}
		},
	}
}

func NewForumFormModel(parent *mainModel, callback func(map[string]string) tea.Cmd) *formModel {
	nameInput := newTextInput("Nome do Fórum")
	descArea := newTextArea("Descrição do Fórum")

	nameInput.Focus()

	fields := []FormField{
		{Name: "Nome", Input: nameInput},
		{Name: "Descrição", Input: descArea},
	}

	return &formModel{
		parent:       parent,
		title:        "Novo Fórum",
		fields:       fields,
		focusIndex:   0,
		submitAction: callback,
	}
}

func (m *formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Em campos de texto de linha única, Enter move para o próximo campo ou submete.
			if _, isTextInput := m.fields[m.focusIndex].Input.(*TextInput); isTextInput {
				if m.focusIndex == len(m.fields)-1 {
					return m, m.submit()
				} else {
					return m, m.nextInput()
				}
			}
		case tea.KeyCtrlS: // Usar Ctrl+S para submeter.
			return m, m.submit()
		case tea.KeyTab, tea.KeyShiftTab:
			if msg.String() == "tab" {
					cmd = m.nextInput()
			} else {
					cmd = m.prevInput()
			}
			return m, cmd
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return navigateBackMsg{} }
		}
	}

	// Atualiza o campo focado
	m.fields[m.focusIndex].Input, cmd = m.fields[m.focusIndex].Input.Update(msg)
	return m, cmd
}

func (m *formModel) submit() tea.Cmd {
	values := make(map[string]string)
	for _, field := range m.fields {
		values[field.Name] = field.Input.Value()
	}
	return m.submitAction(values)
}

func (m *formModel) nextInput() tea.Cmd {
	m.fields[m.focusIndex].Input.Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.fields)
	return m.fields[m.focusIndex].Input.Focus()
}

func (m *formModel) prevInput() tea.Cmd {
	m.fields[m.focusIndex].Input.Blur()
	m.focusIndex--
	if m.focusIndex < 0 {
			m.focusIndex = len(m.fields) - 1
		}
	return m.fields[m.focusIndex].Input.Focus()
}

func (m *formModel) View() string {
	var b strings.Builder
	b.WriteString(m.title + "\n\n")
	for i := range m.fields {
		b.WriteString(m.fields[i].Input.View())
		if i < len(m.fields)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n\n(pressione Ctrl+S para submeter, Esc para cancelar)")
	return b.String()
}
