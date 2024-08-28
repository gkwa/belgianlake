package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const filename = "data.jsonl"

type Item struct {
	Print bool   `json:"print"`
	File  string `json:"file"`
}

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
		}
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	m.shiftPressed = false // Reset shift state after each update
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

func (m model) View() string {
	if m.quitting {
		return "Bye!"
	}
	return baseStyle.Render(m.table.View()) + "\n" +
		"Space: select/deselect row | Shift+Space: select range | Enter: toggle selected rows | u: undo | q: quit"
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func Main() {
	items := loadItems()

	columns := []table.Column{
		{Title: "Print", Width: 5},
		{Title: "File", Width: 80},
	}

	rows := itemsToRows(items)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table:        t,
		items:        items,
		selected:     make(map[int]struct{}),
		lastSelected: -1,
		history:      [][]Item{},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

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

func loadItems() []Item {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var items []Item
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var i Item
		if err := decoder.Decode(&i); err != nil {
			fmt.Println("Error decoding JSON:", err)
			os.Exit(1)
		}
		items = append(items, i)
	}

	return items
}

func saveItems(items []Item) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, i := range items {
		if err := encoder.Encode(i); err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}
	}
}

func saveItemsCmd(items []Item) tea.Cmd {
	return func() tea.Msg {
		saveItems(items)
		return nil
	}
}
