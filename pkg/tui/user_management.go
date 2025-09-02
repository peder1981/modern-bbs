package tui

import (
	"fmt"
	"modern-bbs/internal/database"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)


type userManagementModel struct {
	parent          *mainModel
	users           []database.User
	cursor          int
	keys            *KeyMap
	selectedUser    *database.User
	// State for action selection
	isSelectingAction bool
	actionCursor      int
	actionChoices     []string
	// State for role selection
	isSelectingRole bool
	roleCursor      int
	roleChoices     []string
}

func NewUserManagementModel(parent *mainModel) *userManagementModel {
	return &userManagementModel{
		parent:        parent,
		keys:          DefaultKeyMap,
		actionChoices: []string{"Alterar Papel", "Deletar Usuário", "Resetar Senha"},
		roleChoices:   []string{"user", "moderator", "admin"},
	}
}

func (m *userManagementModel) Init() tea.Cmd {
	return func() tea.Msg {
		users, err := database.GetAllUsers()
		if err != nil {
			return errorMsg{err}
		}
		return usersLoadedMsg{users}
	}
}

func (m *userManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case usersLoadedMsg:
		m.users = msg.users
	case tea.KeyMsg:
			if m.isSelectingRole {
				return m.updateRoleSelection(msg)
			} else if m.isSelectingAction {
				return m.updateActionSelection(msg)
			} else {
				return m.updateUserList(msg)
			}
	}
	return m, nil
}

func (m *userManagementModel) updateUserList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.users)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if len(m.users) > 0 {
			m.selectedUser = &m.users[m.cursor]
			m.isSelectingAction = true
			m.actionCursor = 0 // Reset cursor
		}
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return navigateBackMsg{} }
	}
	return m, nil
}

func (m *userManagementModel) updateActionSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.actionCursor < len(m.actionChoices)-1 {
			m.actionCursor++
		}
	case key.Matches(msg, m.keys.Enter):
		selectedAction := m.actionChoices[m.actionCursor]
		switch selectedAction {
		case "Alterar Papel":
			m.isSelectingAction = false
			m.isSelectingRole = true
			m.roleCursor = 0
		case "Deletar Usuário":
			// Lógica para deletar usuário aqui
			username := m.selectedUser.Username
			err := database.DeleteUser(username)
			m.isSelectingAction = false
			m.selectedUser = nil
			if err != nil {
				return m, func() tea.Msg { return statusMessage{success: false, message: err.Error()} }
			}
			return m, tea.Sequence(m.Init(), func() tea.Msg {
				return statusMessage{success: true, message: fmt.Sprintf("Usuário %s deletado com sucesso", username)}
			})
		case "Resetar Senha":
			// Lógica para resetar senha aqui
			username := m.selectedUser.Username
			err := database.AdminResetPassword(username, "password") // Nova senha padrão
			m.isSelectingAction = false
			m.selectedUser = nil
			if err != nil {
				return m, func() tea.Msg { return statusMessage{success: false, message: err.Error()} }
			}
			return m, tea.Sequence(m.Init(), func() tea.Msg {
				return statusMessage{success: true, message: fmt.Sprintf("Senha de %s resetada para 'password'", username)}
			})
		}
	case key.Matches(msg, m.keys.Back):
		m.isSelectingAction = false
		m.selectedUser = nil
	}
	return m, nil
}

func (m *userManagementModel) updateRoleSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.roleCursor > 0 {
			m.roleCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.roleCursor < len(m.roleChoices)-1 {
			m.roleCursor++
		}
	case key.Matches(msg, m.keys.Enter):
		selectedRole := m.roleChoices[m.roleCursor]
		username := m.selectedUser.Username // Salva o nome de usuário antes de limpar

		err := database.SetUserRole(username, selectedRole)
		m.isSelectingRole = false
		m.selectedUser = nil

		if err != nil {
			return m, func() tea.Msg { return statusMessage{success: false, message: err.Error()} }
		}

		// Recarrega a lista de usuários com mensagem de sucesso
		return m, tea.Sequence(m.Init(), func() tea.Msg {
			return statusMessage{success: true, message: fmt.Sprintf("Papel de %s alterado para %s", username, selectedRole)}
		})
	case key.Matches(msg, m.keys.Back):
		m.isSelectingRole = false
		m.selectedUser = nil
	}
	return m, nil
}

func (m *userManagementModel) View() string {
	if m.isSelectingRole {
		return m.viewRoleSelection()
	} else if m.isSelectingAction {
		return m.viewActionSelection()
	}

	body := "Gerenciamento de Usuários:\n\n"
	for i, user := range m.users {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		body += fmt.Sprintf("%s %s (%s)\n", cursor, user.Username, user.Role)
	}
	return body
}

func (m *userManagementModel) viewActionSelection() string {
	body := fmt.Sprintf("Ações para %s:\n\n", m.selectedUser.Username)
	for i, action := range m.actionChoices {
		cursor := " "
		if m.actionCursor == i {
			cursor = ">"
		}
		body += fmt.Sprintf("%s %s\n", cursor, action)
	}
	return body
}

func (m *userManagementModel) viewRoleSelection() string {
	body := fmt.Sprintf("Alterar papel para %s:\n\n", m.selectedUser.Username)
	for i, role := range m.roleChoices {
		cursor := " "
		if m.roleCursor == i {
			cursor = ">"
		}
		body += fmt.Sprintf("%s %s\n", cursor, role)
	}
	return body
}

func (m *userManagementModel) helpView() string {
	return m.keys.HelpView()
}
