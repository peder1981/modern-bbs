package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// settingsModel representa a visão de configurações.
type settingsModel struct {
	parent  *mainModel
	keys    *KeyMap
	choices []string
	cursor  int
}

// NewSettingsModel cria um novo modelo para a visão de configurações.
func NewSettingsModel(parent *mainModel) *settingsModel {
	m := &settingsModel{
		parent: parent,
		keys:   DefaultKeyMap,
	}

	// Define as opções com base no papel do usuário.
	m.choices = append(m.choices, "Alterar Senha")
	if parent.Role == "moderator" || parent.Role == "admin" {
		m.choices = append(m.choices, "Gerenciar Usuários")
	}
	if parent.Role == "admin" {
		m.choices = append(m.choices, "Criar Novo Usuário")
	}

	return m
}

func (m *settingsModel) Init() tea.Cmd {
	return nil
}

func (m *settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			switch m.choices[m.cursor] {
			case "Alterar Senha":
				m.parent.currentView = formView
				m.parent.formModel = NewChangePasswordFormModel(m.parent)
				return m.parent, m.parent.formModel.Init()
			case "Gerenciar Usuários":
				m.parent.currentView = userManagementView
				m.parent.breadcrumbs = append(m.parent.breadcrumbs, "Gerenciar Usuários")
				m.parent.userManagementModel = NewUserManagementModel(m.parent)
				return m.parent, m.parent.userManagementModel.Init()
			case "Criar Novo Usuário":
				m.parent.currentView = formView
				m.parent.formModel = NewUserFormModel(m.parent)
				return m.parent, m.parent.formModel.Init()
			}
			return m, nil
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return navigateBackMsg{} }
		}
	}
	return m, nil
}

func (m *settingsModel) View() string {
	body := "Selecione uma opção de configuração:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			body += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice))
		} else {
			body += itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice))
		}
		body += "\n"
	}
	return body
}

func (m *settingsModel) helpView() string {
	return fmt.Sprintf("\n  %s • %s • %s • %s",
		m.keys.Up.Help().Key+" "+m.keys.Up.Help().Desc,
		m.keys.Down.Help().Key+" "+m.keys.Down.Help().Desc,
		m.keys.Back.Help().Key+" "+m.keys.Back.Help().Desc,
		m.keys.Quit.Help().Key+" "+m.keys.Quit.Help().Desc,
	)
}
