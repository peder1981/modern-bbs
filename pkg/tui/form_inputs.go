package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TextInput é um wrapper para textinput.Model que implementa formInput.
type TextInput struct{
	textinput.Model
}

func (ti *TextInput) Update(msg tea.Msg) (formInput, tea.Cmd) {
	newModel, cmd := ti.Model.Update(msg)
	ti.Model = newModel
	return ti, cmd
}

// TextArea é um wrapper para textarea.Model que implementa formInput.
type TextArea struct{
	textarea.Model
}

func (ta *TextArea) Update(msg tea.Msg) (formInput, tea.Cmd) {
	newModel, cmd := ta.Model.Update(msg)
	ta.Model = newModel
	return ta, cmd
}
