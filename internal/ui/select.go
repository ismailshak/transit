package ui

import (
	"context"
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Select renders a full-screen list picker built from choices and blocks
// until the user picks one or backs out. On success it returns the matching
// Choice.Key. On cancel (Esc, Ctrl+C, or a cancelled ctx) it returns
// ErrCancelled. On confirming with nothing matched returns ErrNoSelection.
func Select(ctx context.Context, title string, choices []Choice) (string, error) {
	listItems := convertListItems(choices)
	keys := &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}

	d := newItemDelegate(keys)
	l := list.New(listItems, d, 0, 0)
	l.Title = title

	m, err := tea.NewProgram(
		listModel{list: l},
		tea.WithContext(ctx),
		tea.WithAltScreen(),
	).Run()

	if errors.Is(err, context.Canceled) {
		return "", ErrCancelled
	}

	if err != nil {
		return "", err
	}

	lm := m.(listModel)
	if lm.cancelled {
		return "", ErrCancelled
	}

	if lm.selectedKey == "" {
		return "", ErrNoSelection
	}

	return lm.selectedKey, nil
}

// list.New wants []list.Item; wrap each Choice so it satisfies that
// interface without Choice itself knowing bubbles exists.
func convertListItems(choices []Choice) []list.Item {
	items := make([]list.Item, len(choices))
	for i, c := range choices {
		items[i] = choiceItem{c}
	}
	return items
}

// -- Internal model for the list component --

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type listModel struct {
	list        list.Model
	selectedKey string
	// cancelled distinguishes "user backed out" from "confirmed with nothing
	// selected" — both leave selectedKey empty, but only the former should
	// produce ErrCancelled.
	cancelled bool
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "q":
			// While actively typing a filter it's a literal character,
			// so don't steal it from the list's own filter input.
			if m.list.FilterState() != list.Filtering {
				m.cancelled = true
				return m, tea.Quit
			}
		case "esc":
			// Esc is overloaded in bubbles. With an applied filter it clears
			// the filter first, so only treat it as cancel when there's no filter to clear.
			// Falling through lets the list handle both the "clear filter" and "still typing"
			// cases itself.
			if m.list.FilterState() == list.Unfiltered {
				m.cancelled = true
				return m, tea.Quit
			}
		case "enter":
			selected := m.list.SelectedItem()
			// If filtering and nothing matched, hitting enter will return nil
			if selected != nil {
				m.selectedKey = selected.(choiceItem).Key()
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

// -- Custom delegate to handle list items --

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

// Additional short help entries. This satisfies the help.KeyMap interface.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}
