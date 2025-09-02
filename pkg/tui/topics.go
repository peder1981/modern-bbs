package tui

import (
	"fmt"
	"modern-bbs/internal/database"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// topicsModel representa a visão da lista de tópicos de um fórum.
type topicsModel struct {
	keys          *KeyMap
	parent        *mainModel
	forum         *database.Forum
	topics        []*database.Topic
	cursor        int
	quitting      bool
	navToPosts       *database.Topic // Tópico para o qual navegar
	creatingTopic    bool            // Sinaliza se estamos criando um novo tópico
	confirmingDelete bool
}

type topicsLoadedMsg struct {
	topics []*database.Topic
	err    error
}

type reloadTopicsMsg struct{}

// NewTopicsModel cria um novo modelo para a visão de tópicos.
func NewTopicsModel(parent *mainModel, forum *database.Forum) *topicsModel {
	return &topicsModel{
		keys:   DefaultKeyMap,
		parent: parent,
		forum:  forum,
	}
}

func (m *topicsModel) Init() tea.Cmd {
	return func() tea.Msg {
		topics, err := database.GetTopicsByForumID(int(m.forum.ID))
		return topicsLoadedMsg{topics: topics, err: err}
	}
}

func (m *topicsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case topicsLoadedMsg:
		if msg.err != nil {
			// TODO: Tratar o erro de forma mais elegante
			return m, tea.Quit
		}
		m.topics = msg.topics
		return m, nil
	case reloadTopicsMsg:
		return m, m.Init()
	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "s", "S":
				if len(m.topics) > 0 {
					selectedTopicID := m.topics[m.cursor].ID
					cmd := func() tea.Msg {
						err := database.DeleteTopic(selectedTopicID)
						if err != nil {
							return statusMessage{success: false, message: "Erro ao deletar tópico: " + err.Error()}
						}
						return reloadTopicsMsg{}
					}
					m.confirmingDelete = false
					return m, cmd
				}
			case "n", "N":
				m.confirmingDelete = false
				return m, nil
			}
			return m, nil // Não processa outras teclas durante a confirmação
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.topics)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.topics) > 0 {
				m.navToPosts = m.topics[m.cursor]
				return m, nil // Retorna para o mainModel que irá lidar com a navegação
			}
		case key.Matches(msg, m.keys.New):
			if m.parent.Role == "admin" || m.parent.Role == "moderator" {
				m.creatingTopic = true
				return m, nil // O mainModel irá lidar com a transição para o formulário
			}
		case key.Matches(msg, m.keys.Delete):
			if (m.parent.Role == "admin" || m.parent.Role == "moderator") && len(m.topics) > 0 {
				m.confirmingDelete = true
			}
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return navigateBackMsg{} }
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *topicsModel) View() string {
	if m.quitting {
		return ""
	}

	header := headerStyle.Render(fmt.Sprintf("Tópicos em '%s'", m.forum.Name))

	body := ""
	if len(m.topics) == 0 {
		body = "Nenhum tópico encontrado.\n"
	} else {
		for i, topic := range m.topics {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				body += selectedItemStyle.Render(fmt.Sprintf("%s %s (por %s)", cursor, topic.Title, topic.Username))
			} else {
				body += itemStyle.Render(fmt.Sprintf("%s %s (por %s)", cursor, topic.Title, topic.Username))
			}
			body += "\n"
		}
	}

	footer := footerStyle.Render(m.helpView())

	if m.confirmingDelete && len(m.topics) > 0 {
		topicTitle := m.topics[m.cursor].Title
		body += fmt.Sprintf("\n\nTem certeza que deseja deletar o tópico '%s'? (s/n)", topicTitle)
	}

	return fmt.Sprintf("%s\n\n%s\n%s", header, body, footer)
}

func (m *topicsModel) helpView() string {
	help := []string{
		m.keys.Up.Help().Key + " " + m.keys.Up.Help().Desc,
		m.keys.Down.Help().Key + " " + m.keys.Down.Help().Desc,
	}

	if m.parent.Role == "admin" || m.parent.Role == "moderator" {
		help = append(help, m.keys.New.Help().Key+" "+m.keys.New.Help().Desc)
		help = append(help, m.keys.Delete.Help().Key+" "+m.keys.Delete.Help().Desc)
	}

	help = append(help, m.keys.Back.Help().Key+" "+m.keys.Back.Help().Desc)
	help = append(help, m.keys.Quit.Help().Key+" "+m.keys.Quit.Help().Desc)

	return "\n  " + strings.Join(help, " • ")
}
