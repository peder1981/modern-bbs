package tui

import (
	"fmt"
	"modern-bbs/internal/database"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)


// forumsModel representa a visão da lista de fóruns.
type forumsModel struct {
	parent      *mainModel
	forums      []database.Forum
	cursor      int
	quitting    bool
	navToTopics *database.Forum // Fórum selecionado para navegação
	keys        *KeyMap
}

// NewForumsModel cria um novo modelo para a visão de fóruns.
func NewForumsModel(parent *mainModel) *forumsModel {
	return &forumsModel{
		parent: parent,
		keys:   DefaultKeyMap,
	}
}

func (m *forumsModel) Init() tea.Cmd {
	m.parent.isLoading = true
	return m.loadForumsCmd
}

// loadForumsCmd é um comando que carrega os fóruns do banco de dados.
func (m *forumsModel) loadForumsCmd() tea.Msg {
	forums, err := database.GetAllForums()
	if err != nil {
		return errorMsg{err}
	}
	return forumsLoadedMsg{forums}
}

func (m *forumsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case forumsLoadedMsg:
		m.parent.isLoading = false
		m.forums = msg.forums
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.forums)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.forums) > 0 {
				m.navToTopics = &m.forums[m.cursor]
				return m, nil // Retorna para o mainModel que irá lidar com a navegação
			}
		case key.Matches(msg, m.keys.Back):
			// Envia uma mensagem para o mainModel para navegar para trás.
			return m, func() tea.Msg { return navigateBackMsg{} }
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *forumsModel) View() string {
	if m.quitting {
		return ""
	}

	body := ""
	for i, forum := range m.forums {
		cursor := " " // Espaço em branco para o cursor não selecionado
		if m.cursor == i {
			cursor = ">" // Cursor para o item selecionado
			body += selectedItemStyle.Render(fmt.Sprintf("%s %s", cursor, forum.Name))
		} else {
			body += itemStyle.Render(fmt.Sprintf("%s %s", cursor, forum.Name))
		}
		body += "\n"
	}

	return body
}

func (m *forumsModel) helpView() string {
	return m.keys.HelpView()
}
