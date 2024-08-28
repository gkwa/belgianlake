package core

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table        table.Model
	items        []Item
	quitting     bool
	selected     map[int]struct{}
	shiftPressed bool
	lastSelected int
	history      [][]Item
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "shift":
			m.shiftPressed = true
		case " ":
			cursor := m.table.Cursor()
			if m.shiftPressed {
				m.selectRange(cursor)
			} else {
				m.toggleSelection(cursor)
				if cursor < len(m.items)-1 {
					m.moveToNextRow()
				}
			}
			m.lastSelected = cursor
			m.updateTableRows()
			return m, nil
		case "enter":
			m.saveState()
			for index := range m.selected {
				m.items[index].Print = !m.items[index].Print
			}
			m.selected = make(map[int]struct{})
			m.updateTableRows()
			return m, saveItemsCmd(m.items)
		case "u":
			if len(m.history) > 0 {
				m.items = m.history[len(m.history)-1]
				m.history = m.history[:len(m.history)-1]
				m.selected = make(map[int]struct{})
				m.updateTableRows()
				return m, saveItemsCmd(m.items)
			}
		case "t":
			m.saveState()
			for i := range m.items {
				m.items[i].Print = !m.items[i].Print
			}
			m.updateTableRows()
			return m, saveItemsCmd(m.items)
		case "a":
			for i := range m.items {
				m.selected[i] = struct{}{}
			}
			m.updateTableRows()
			return m, nil
		case "d":
			m.selected = make(map[int]struct{})
			m.updateTableRows()
			return m, nil
		case "e":
			m.saveState()
			allEnabled := true
			for _, item := range m.items {
				if !item.Print {
					allEnabled = false
					break
				}
			}
			for i := range m.items {
				m.items[i].Print = !allEnabled
			}
			m.updateTableRows()
			return m, saveItemsCmd(m.items)
		case "x":
			m.saveState()
			cursor := m.table.Cursor()
			m.items[cursor].Print = !m.items[cursor].Print
			if cursor < len(m.items)-1 {
				m.moveToNextRow()
			}
			m.updateTableRows()
			return m, saveItemsCmd(m.items)
		}
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	m.shiftPressed = false
	return m, cmd
}

func (m *model) saveState() {
	itemsCopy := make([]Item, len(m.items))
	copy(itemsCopy, m.items)
	m.history = append(m.history, itemsCopy)
}

func (m *model) toggleSelection(index int) {
	if _, ok := m.selected[index]; ok {
		delete(m.selected, index)
	} else {
		m.selected[index] = struct{}{}
	}
}

func (m *model) selectRange(endIndex int) {
	startIndex := m.lastSelected
	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}
	for i := startIndex; i <= endIndex; i++ {
		m.selected[i] = struct{}{}
	}
}

func (m *model) moveToNextRow() {
	nextRow := (m.table.Cursor() + 1) % len(m.items)
	m.table.SetCursor(nextRow)
}

func (m model) View() string {
	if m.quitting {
		return "Bye!"
	}
	return baseStyle.Render(m.table.View()) + "\n" +
		"Space: select/deselect row | Shift+Space: select range | Enter: toggle selected rows\n" +
		"t: toggle all | a: select all | d: deselect all | e: toggle enable/disable all | x: toggle current row | u: undo | q: quit"
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m *model) updateTableRows() {
	rows := make([]table.Row, len(m.items))
	for i, item := range m.items {
		printStatus := "[ ]"
		if item.Print {
			printStatus = "[x]"
		}
		if _, ok := m.selected[i]; ok {
			printStatus = ">" + printStatus
		}
		rows[i] = table.Row{printStatus, item.File}
	}
	m.table.SetRows(rows)
}

func itemsToRows(items []Item) []table.Row {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		printStatus := "[ ]"
		if item.Print {
			printStatus = "[x]"
		}
		rows[i] = table.Row{printStatus, item.File}
	}
	return rows
}
