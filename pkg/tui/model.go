package tui

import (
	"fmt"
	"modern-bbs/internal/database"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("77"))
	selectedItemStyle     = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Background(lipgloss.Color("57"))
	itemStyle             = lipgloss.NewStyle().PaddingLeft(2)
	footerStyle           = lipgloss.NewStyle().Faint(true)
	statusMessageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // Verde
	errorStatusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Vermelho
)

type view int

const (
	mainMenuView view = iota
	forumsView
	topicsView
	postsView
	formView
	settingsView
	userManagementView
	forumManagementView
	adminView
)

// Mensagens para comunicação entre modelos e para operações assíncronas.
type statusMessageTimeoutMsg struct{}

type statusMessage struct {
	success bool
	message string
}
type errorMsg struct{ err error }

func (e errorMsg) Error() string { return e.err.Error() }

// Mensagens para sinalizar o carregamento de dados.
type forumsLoadedMsg struct{ forums []database.Forum }

// Esta mensagem pode ser necessária para o userManagementModel
type usersLoadedMsg struct{ users []database.User }

// Mensagens para sinalizar a conclusão de ações do formulário.
type passwordUpdatedMsg struct{}
type topicCreatedMsg struct{ forum *database.Forum }
type postCreatedMsg struct{ topic *database.Topic }
type userCreatedMsg struct{}
type navigateBackMsg struct{}

// mainModel é o modelo principal que gerencia as visões da aplicação.
type mainModel struct {
	User                string
	Role                string // 'user', 'moderator', 'admin'
	currentView         view
	Choices             []string
	Cursor              int
	forumsModel         *forumsModel
	topicsModel         *topicsModel
	postsModel          *postsModel
	formModel           *formModel
	settingsModel       *settingsModel
	userManagementModel *userManagementModel
	forumManagementModel *forumManagementModel
	adminModel          *adminModel

	// UX Enhancements
	spinner       spinner.Model
	isLoading     bool
	statusMessage string
	breadcrumbs   []string
}

// InitialModel cria o nosso modelo inicial com o nome e o papel do usuário.
func InitialModel(user, role string) *mainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	choices := []string{"Ver Fóruns", "Configurações"}
	if role == "admin" {
		choices = append(choices, "Administração")
	}
	choices = append(choices, "Sair")

	return &mainModel{
		User:        user,
		Role:        role,
		currentView: mainMenuView,
		Choices:     choices,
		spinner:     s,
		isLoading:   false,
		breadcrumbs: []string{"Home"},
	}
}

// Init é a primeira função que é executada quando o programa inicia.
func (m *mainModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update lida com as entradas do usuário e atualiza o estado.
func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Comandos globais, independentemente da view

	case statusMessageTimeoutMsg:
		m.statusMessage = ""
		return m, nil
	case errorMsg:
		m.isLoading = false
		m.statusMessage = "Erro: " + msg.err.Error()
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return statusMessageTimeoutMsg{}
		})
	case forumsLoadedMsg:
		var newModel tea.Model
		newModel, cmd = m.forumsModel.Update(msg)
		m.forumsModel = newModel.(*forumsModel)
		return m, cmd
	case usersLoadedMsg:
		var newModel tea.Model
		newModel, cmd = m.userManagementModel.Update(msg)
		m.userManagementModel = newModel.(*userManagementModel)
		return m, cmd
	case passwordUpdatedMsg:
		m.statusMessage = "Senha alterada com sucesso!"
		m.currentView = settingsView
		return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg { return statusMessageTimeoutMsg{} })
	case topicCreatedMsg:
		m.statusMessage = "Tópico criado com sucesso!"
		m.currentView = topicsView
		tm := NewTopicsModel(m, msg.forum)
		m.topicsModel = tm
		return m, tm.Init()
	case postCreatedMsg:
		m.statusMessage = "Post criado com sucesso!"
		m.currentView = postsView
		pm := NewPostsModel(m, msg.topic)
		m.postsModel = pm
		return m, pm.Init()
	case statusMessage:
		if msg.success {
			m.statusMessage = msg.message
		} else {
			m.statusMessage = "Erro: " + msg.message
		}
		m.currentView = settingsView // Volta para a tela de configurações após a ação
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return statusMessageTimeoutMsg{} })
	case navigateBackMsg:
		if len(m.breadcrumbs) > 1 {
			m.breadcrumbs = m.breadcrumbs[:len(m.breadcrumbs)-1]
			newView := m.breadcrumbs[len(m.breadcrumbs)-1]

			switch newView {
			case "Home":
				m.currentView = mainMenuView
			case "Administração":
				m.currentView = adminView
			case "Fóruns":
				m.currentView = forumsView
			case "Tópicos":
				m.currentView = topicsView
			case "Configurações":
				m.currentView = settingsView
			}
		}
		return m, nil
	}

	var newModel tea.Model

	switch m.currentView {
	case formView:
		newModel, cmd = m.formModel.Update(msg)
		m.formModel = newModel.(*formModel)
	case postsView:
		newModel, cmd = m.postsModel.Update(msg)
		m.postsModel = newModel.(*postsModel)
	case topicsView:
		newModel, cmd = m.topicsModel.Update(msg)
		m.topicsModel = newModel.(*topicsModel)
	case forumsView:
		newModel, cmd = m.forumsModel.Update(msg)
		m.forumsModel = newModel.(*forumsModel)
	case adminView:
		newModel, cmd = m.adminModel.Update(msg)
		if _, ok := newModel.(*adminModel); ok {
			m.adminModel = newModel.(*adminModel)
		} else {
			// Se o modelo retornado não for adminModel, significa que estamos voltando ao menu principal.
			return newModel, cmd
		}
	case userManagementView:
		newModel, cmd = m.userManagementModel.Update(msg)
		m.userManagementModel = newModel.(*userManagementModel)
	case forumManagementView:
		newModel, cmd = m.forumManagementModel.Update(msg)
		m.forumManagementModel = newModel.(*forumManagementModel)
	case settingsView:
		newModel, cmd = m.settingsModel.Update(msg)
		m.settingsModel = newModel.(*settingsModel)
	default: // mainMenuView
		return m.updateMainMenu(msg)
	}

	// Lógica de navegação para frente
	if m.adminModel != nil && m.adminModel.navigateToUserManagement {
		m.currentView = userManagementView
		m.breadcrumbs = append(m.breadcrumbs, "Gerenciamento de Usuários")
		if m.userManagementModel == nil {
			m.userManagementModel = NewUserManagementModel(m)
			m.forumManagementModel = NewForumManagementModel(m)
		}
		cmd = m.userManagementModel.Init()
		m.adminModel.navigateToUserManagement = false
	} else if m.adminModel != nil && m.adminModel.navigateToForumManagement {
		m.currentView = forumManagementView
		m.breadcrumbs = append(m.breadcrumbs, "Gerenciamento de Fóruns")
		if m.forumManagementModel == nil {
			m.forumManagementModel = NewForumManagementModel(m)
		}
		cmd = m.forumManagementModel.Init()
		m.adminModel.navigateToForumManagement = false
	} else if m.forumManagementModel != nil && m.forumManagementModel.navigateToForm {
		m.currentView = formView
		m.breadcrumbs = append(m.breadcrumbs, "Novo Fórum")

		callback := func(values map[string]string) tea.Cmd {
			return func() tea.Msg {
				_, err := database.CreateForum(values["Nome"], values["Descrição"])
				if err != nil {
					return statusMessage{success: false, message: "Erro ao criar fórum: " + err.Error()}
				}
				return statusMessage{success: true, message: fmt.Sprintf("Fórum '%s' criado com sucesso!", values["Nome"])}
			}
		}

		m.formModel = NewForumFormModel(m, callback)
		cmd = m.formModel.Init()
		m.forumManagementModel.navigateToForm = false
	} else if m.forumManagementModel != nil && m.forumManagementModel.navigateToEditForm {
		m.currentView = formView
		m.breadcrumbs = append(m.breadcrumbs, "Editando Fórum")
		m.formModel = NewEditForumFormModel(m, m.forumManagementModel.selectedForum)
		cmd = m.formModel.Init()
		m.forumManagementModel.navigateToEditForm = false
	} else if m.forumsModel != nil && m.forumsModel.navToTopics != nil {
		m.currentView = topicsView
		m.breadcrumbs = append(m.breadcrumbs, m.forumsModel.navToTopics.Name)
		m.topicsModel = NewTopicsModel(m, m.forumsModel.navToTopics)
		cmd = m.topicsModel.Init()
		m.forumsModel.navToTopics = nil
	} else if m.topicsModel != nil && m.topicsModel.navToPosts != nil {
		m.currentView = postsView
		m.breadcrumbs = append(m.breadcrumbs, m.topicsModel.navToPosts.Title)
		m.postsModel = NewPostsModel(m, m.topicsModel.navToPosts)
		cmd = m.postsModel.Init()
		m.topicsModel.navToPosts = nil
	} else if m.topicsModel != nil && m.topicsModel.creatingTopic {
		m.currentView = formView
		m.breadcrumbs = append(m.breadcrumbs, "Novo Tópico")
		m.formModel = NewTopicFormModel(m, m.topicsModel.forum)
		cmd = m.formModel.Init()
		m.topicsModel.creatingTopic = false
	} else if m.postsModel != nil && m.postsModel.creatingPost {
		m.currentView = formView
		m.breadcrumbs = append(m.breadcrumbs, "Novo Post")
		m.formModel = NewPostFormModel(m, m.postsModel.topic)
		cmd = m.formModel.Init()
		m.postsModel.creatingPost = false
	}

	return m, cmd
}

// updateMainMenu lida com a lógica de atualização do menu principal.
func (m *mainModel) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter":
			switch m.Choices[m.Cursor] {
			case "Ver Fóruns":
				m.currentView = forumsView
				m.breadcrumbs = []string{"Home", "Fóruns"}
				if m.forumsModel == nil {
					m.forumsModel = NewForumsModel(m)
				}
				return m, m.forumsModel.Init()
			case "Configurações":
				m.currentView = settingsView
				m.breadcrumbs = []string{"Home", "Configurações"}
				if m.settingsModel == nil {
					m.settingsModel = NewSettingsModel(m)
				}
				return m, m.settingsModel.Init()
			case "Administração":
				m.currentView = adminView
				m.breadcrumbs = []string{"Home", "Administração"}
				if m.adminModel == nil {
					m.adminModel = NewAdminModel(m)
				}
				return m, m.adminModel.Init()
			case "Sair":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// View renderiza a UI.
func (m *mainModel) View() string {
	// Renderiza o cabeçalho com breadcrumbs
	header := m.renderHeader()

	// Renderiza a view atual
	var currentViewContent string
	switch m.currentView {
	case mainMenuView:
		currentViewContent = m.viewMainMenu()
	case forumsView:
		currentViewContent = m.forumsModel.View()
	case topicsView:
		currentViewContent = m.topicsModel.View()
	case postsView:
		currentViewContent = m.postsModel.View()
	case formView:
		currentViewContent = m.formModel.View()
	case settingsView:
		currentViewContent = m.settingsModel.View()
	case userManagementView:
		currentViewContent = m.userManagementModel.View()
	case forumManagementView:
		currentViewContent = m.forumManagementModel.View()
	case adminView:
		currentViewContent = m.adminModel.View()
	}

	// Renderiza o rodapé
	footer := m.renderFooter()

	// Adiciona a mensagem de status se houver uma
	statusMsg := ""
	if m.statusMessage != "" {
		style := statusMessageStyle
		if strings.HasPrefix(m.statusMessage, "Erro") {
			style = errorStatusMessageStyle
		}
		statusMsg = style.Render(m.statusMessage)
	}

	// Junta tudo
	return fmt.Sprintf("%s\n\n%s\n%s\n%s", header, currentViewContent, statusMsg, footer)
}

// renderHeader renderiza o cabeçalho da UI.
func (m *mainModel) renderHeader() string {
	return headerStyle.Render(strings.Join(m.breadcrumbs, " > "))
}

// renderFooter renderiza o rodapé da UI.
func (m *mainModel) renderFooter() string {
	var help string
	switch m.currentView {
	case forumsView:
		help = m.forumsModel.helpView()
	case topicsView:
		// help = m.topicsModel.helpView()
	case postsView:
		// help = m.postsModel.helpView()
	case formView:
		// help = m.formModel.helpView()
	case settingsView:
		help = m.settingsModel.helpView()
	case userManagementView:
		help = m.userManagementModel.helpView()
	case forumManagementView:
		// help = m.forumManagementModel.helpView()
	default:
		help = "Use as setas para navegar e 'enter' para selecionar. Pressione 'q' para sair."
	}

	return footerStyle.Render(help)
}

// viewMainMenu renderiza a UI do menu principal.
func (m *mainModel) viewMainMenu() string {
	s := fmt.Sprintf("Bem-vindo ao Modern BBS, %s!\n\n", m.User)

	for i, choice := range m.Choices {
		if m.Cursor == i {
			s += selectedItemStyle.Render(fmt.Sprintf("> %s", choice))
		} else {
			s += itemStyle.Render(fmt.Sprintf("  %s", choice))
		}
		s += "\n"
	}

	return s
}
