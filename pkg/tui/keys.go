package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap define um conjunto de atalhos de teclado para a aplicação.
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
	New   key.Binding // Para criar novos itens (tópicos/posts)
	Delete key.Binding
}

// DefaultKeyMap é a instância global dos atalhos de teclado.
var DefaultKeyMap = &KeyMap{
	Up: key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "mover para cima")),
	Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "mover para baixo")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "selecionar")),
	Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "voltar")),
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "sair")),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "novo"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deletar"),
	),
}

// HelpView retorna uma string com a ajuda dos atalhos de teclado.
func (k *KeyMap) HelpView() string {
	return fmt.Sprintf("%s: %s, %s: %s, %s: %s, %s: %s, %s: %s, %s: %s, %s: %s",
		k.Up.Help().Key, k.Up.Help().Desc,
		k.Down.Help().Key, k.Down.Help().Desc,
		k.Enter.Help().Key, k.Enter.Help().Desc,
		k.Back.Help().Key, k.Back.Help().Desc,
		k.New.Help().Key, k.New.Help().Desc,
		k.Quit.Help().Key, k.Quit.Help().Desc,
	)
}
