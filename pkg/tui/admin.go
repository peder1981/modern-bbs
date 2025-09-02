package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var adminTitleStyle = lipgloss.NewStyle().MarginLeft(2)

type adminMenuItem struct {
	title, desc string
}

func (i adminMenuItem) Title() string       { return i.title }
func (i adminMenuItem) Description() string { return i.desc }
func (i adminMenuItem) FilterValue() string { return i.title }

// adminModel gerencia a tela de administração.
type adminModel struct {
	main                       *mainModel
	list                       list.Model
	navigateToUserManagement   bool
	navigateToForumManagement  bool
}

// NewAdminModel cria um novo modelo para a tela de administração.
func NewAdminModel(main *mainModel) *adminModel {
	items := []list.Item{
		adminMenuItem{title: "Gerenciamento de Usuários", desc: "Editar, deletar e alterar papéis de usuários"},
		adminMenuItem{title: "Gerenciamento de Fóruns", desc: "Criar, editar e deletar fóruns"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Menu de Administração"

	return &adminModel{main: main, list: l}
}

// Init inicializa o modelo.
func (m *adminModel) Init() tea.Cmd {
	return nil
}

// Update lida com as mensagens e atualiza o modelo.
func (m *adminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.main, func() tea.Msg { return navigateBackMsg{} }
		case "enter":
			i, ok := m.list.SelectedItem().(adminMenuItem)
			if ok {
				switch i.title {
				case "Gerenciamento de Usuários":
					m.navigateToUserManagement = true
				case "Gerenciamento de Fóruns":
					m.navigateToForumManagement = true
					return m, nil
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renderiza a tela de administração.
func (m *adminModel) View() string {
	return adminTitleStyle.Render(m.list.View())
}
