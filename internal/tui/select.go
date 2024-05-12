package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Select struct {
	list listModel
}

// Convenience interface for items in the list so I can access Key/Title/Description
type Item interface {
	// Key is the unique identifier for the item and will be the value returned when the user selects an item
	Key() string
	// Title is the primary display value for the item
	Title() string
	// Description is the secondary display value for the item that adds more context
	Description() string
	// FilterValue is the value that will be used for fuzzy matching when filtering the list
	FilterValue() string
}

func NewSelectPrompt[T Item](title string, items []T) *Select {
	listItems := convertListItems(items)
	keys := &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}

	d := newItemDelegate(keys)
	l := list.New(listItems, d, 0, 0)
	l.Title = title

	listModel := listModel{
		list: l,
	}

	return &Select{
		list: listModel,
	}
}

func convertListItems[T list.Item](items []T) []list.Item {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	return listItems
}

// Render the list and block until the user exits or selects an item.
// Returns the `Key` of the selected item or empty string if the user canceled
func (l *Select) Render() string {
	m, _ := tea.NewProgram(l.list, tea.WithAltScreen()).Run() // TODO: handle error
	return m.(listModel).selectedKey
}

// Internal model for the list component

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type listModel struct {
	list        list.Model
	selectedKey string
	canceled    bool
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			selected := m.list.SelectedItem()
			// If filtering and nothing matched, hitting enter will return nil
			if selected != nil {
				m.selectedKey = selected.(Item).Key()
			}

			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return docStyle.Render(m.list.View())
}

// Custom delegate to handle list items

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	help := []key.Binding{keys.choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}
