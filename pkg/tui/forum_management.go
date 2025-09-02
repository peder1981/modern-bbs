package tui

import (
	"fmt"
	"modern-bbs/internal/database"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// forumManagementModel gerencia a tela de gerenciamento de fóruns.
type forumManagementModel struct {
	parent           *mainModel
	forums           []database.Forum
	cursor           int
	keys             *KeyMap
	navigateToForm   bool
	navigateToEditForm bool
	selectedForum    *database.Forum
	confirmingDelete bool
}

// NewForumManagementModel cria um novo modelo para a tela de gerenciamento de fóruns.
func NewForumManagementModel(parent *mainModel) *forumManagementModel {
	return &forumManagementModel{
		parent: parent,
		keys:   DefaultKeyMap,
	}
}

// Init carrega os fóruns do banco de dados.
func (m *forumManagementModel) Init() tea.Cmd {
	return func() tea.Msg {
		forums, err := database.GetAllForums()
		if err != nil {
			return errorMsg{err}
		}
		return forumsLoadedMsg{forums}
	}
}

// Update processa as mensagens para o modelo.
func (m *forumManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case forumsLoadedMsg:
		m.forums = msg.forums
	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "s", "S":
				if len(m.forums) > 0 {
					selectedForumID := m.forums[m.cursor].ID
					cmd := func() tea.Msg {
						err := database.DeleteForum(selectedForumID)
						if err != nil {
							return statusMessage{success: false, message: "Erro ao deletar fórum: " + err.Error()}
						}
						// Recarrega a lista de fóruns após a deleção
						return m.Init()()
					}
					m.confirmingDelete = false
					return m, cmd
				}
			case "n", "N":
				m.confirmingDelete = false
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.forums)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.New):
			m.navigateToForm = true
			return m, nil
		case msg.String() == "e": // Editar fórum
			if len(m.forums) > 0 {
				m.selectedForum = &m.forums[m.cursor]
				m.navigateToEditForm = true
			}
			return m, nil

		case key.Matches(msg, m.keys.Delete):
			if len(m.forums) > 0 {
				m.confirmingDelete = true
			}
			return m, nil

		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return navigateBackMsg{} }
		}
	}
	return m, nil
}

// View renderiza a tela de gerenciamento de fóruns.
func (m *forumManagementModel) View() string {
	body := "Gerenciamento de Fóruns:\n\n"
	for i, forum := range m.forums {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		body += fmt.Sprintf("%s %s\n", cursor, forum.Name)
	}

	if m.confirmingDelete && len(m.forums) > 0 {
		forumName := m.forums[m.cursor].Name
		body += fmt.Sprintf("\n\nTem certeza que deseja deletar o fórum '%s'? (s/n)", forumName)
	}

	return body
}
