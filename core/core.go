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
	table    table.Model
	items    []Item
	quitting bool
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
		case " ":
			index := m.table.Cursor()
			m.items[index].Print = !m.items[index].Print
			m.updateTableRows()
			return m, saveItemsCmd(m.items)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Bye!"
	}
	return baseStyle.Render(m.table.View()) + "\n"
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func Main() {
	items := loadItems()

	columns := []table.Column{
		{Title: "Print", Width: 5},
		{Title: "File", Width: 50},
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
		table: t,
		items: items,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m *model) updateTableRows() {
	m.table.SetRows(itemsToRows(m.items))
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
