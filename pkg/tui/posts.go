package tui

import (
	"fmt"
	"modern-bbs/internal/database"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// postsModel representa a visão dos posts de um tópico.
type postsModel struct {
	keys             *KeyMap
	parent           *mainModel
	topic            *database.Topic
	posts            []*database.Post
	cursor           int
	quitting         bool
	creatingPost     bool // Sinaliza se estamos criando um novo post (resposta)
	confirmingDelete bool
}

type postsLoadedMsg struct {
	posts []*database.Post
	err   error
}

type reloadPostsMsg struct{}

// NewPostsModel cria um novo modelo para a visão de posts.
func NewPostsModel(parent *mainModel, topic *database.Topic) *postsModel {
	return &postsModel{
		keys:   DefaultKeyMap,
		parent: parent,
		topic:  topic,
	}
}

func (m *postsModel) Init() tea.Cmd {
	return func() tea.Msg {
		posts, err := database.GetPostsByTopicID(m.topic.ID)
		return postsLoadedMsg{posts: posts, err: err}
	}
}

func (m *postsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case postsLoadedMsg:
		if msg.err != nil {
			return m, tea.Quit // Tratar erro
		}
		m.posts = msg.posts
		return m, nil
	case reloadPostsMsg:
		return m, m.Init()
	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "s", "S":
				if len(m.posts) > 0 {
					selectedPostID := m.posts[m.cursor].ID
					cmd := func() tea.Msg {
						err := database.DeletePost(selectedPostID)
						if err != nil {
							return statusMessage{success: false, message: "Erro ao deletar post: " + err.Error()}
						}
						return reloadPostsMsg{}
					}
					m.confirmingDelete = false
					return m, cmd
				}
			case "n", "N":
				m.confirmingDelete = false
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.posts)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.New):
			if m.parent.Role != "" {
				m.creatingPost = true
				return m, nil
			}
		case key.Matches(msg, m.keys.Delete):
			if (m.parent.Role == "admin" || m.parent.Role == "moderator") && len(m.posts) > 0 {
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

func (m *postsModel) View() string {
	if m.quitting {
		return ""
	}

	if m.confirmingDelete {
		return fmt.Sprintf("Você tem certeza que quer deletar o post selecionado?\n\n(s/n)")
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("Lendo: %s", m.topic.Title)) + "\n\n")

	if len(m.posts) == 0 {
		b.WriteString("Nenhuma postagem neste tópico ainda.")
	} else {
		for i, post := range m.posts {
			style := itemStyle
			if i == m.cursor {
				style = selectedItemStyle
			}
			authorLine := fmt.Sprintf("De: %s em %s", post.Username, post.CreatedAt.Format(time.RFC822))
			b.WriteString(style.Render(authorLine))
			b.WriteString("\n")
			b.WriteString(style.Render(post.Content))
			b.WriteString("\n---\n")
		}
	}

	b.WriteString("\n" + footerStyle.Render(m.helpView()))

	return b.String()
}

func (m *postsModel) helpView() string {
	help := []string{
		m.keys.Up.Help().Key + "/" + m.keys.Down.Help().Key + " navegar",
	}

	if m.parent.Role != "" {
		help = append(help, m.keys.New.Help().Key+" "+m.keys.New.Help().Desc)
	}

	if m.parent.Role == "admin" || m.parent.Role == "moderator" {
		help = append(help, m.keys.Delete.Help().Key+" "+m.keys.Delete.Help().Desc)
	}

	help = append(help, m.keys.Back.Help().Key+" "+m.keys.Back.Help().Desc)
	help = append(help, m.keys.Quit.Help().Key+" "+m.keys.Quit.Help().Desc)

	return strings.Join(help, " • ")
}
