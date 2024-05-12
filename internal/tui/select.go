package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type List struct {
	list listModel
}

// TODO: Not enforced
// Convenience interface for items in the list so I can access Key/Title/Description
type Item interface {
	Key() string
	Title() string
	Description() string
	FilterValue() string
}

func NewListPrompt(title string, items []list.Item) *List {
	keys := &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}

	d := newItemDelegate(keys)
	l := list.New(items, d, 0, 0)
	l.Title = title

	listModel := listModel{
		list: l,
	}

	return &List{
		list: listModel,
	}
}

func ToListItems[T list.Item](items []T) []list.Item {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	return listItems
}

// Render the list and block until the user exits or selects an item
// Returns the index of the selected item
func (l *List) Render() string {
	m, _ := tea.NewProgram(l.list, tea.WithAltScreen()).Run() // TODO: handle error
	return m.(listModel).selectedKey
}

//
// Internal model for the list component
//

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
