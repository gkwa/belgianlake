package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-logr/logr"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Hello(logger logr.Logger) {
	logger.V(1).Info("Debug: Entering Hello function")
	logger.Info("Hello, World!")
	logger.V(1).Info("Debug: Exiting Hello function")
}

const filename = "data.jsonl"

type Item struct {
	Print bool   `json:"print"`
	File  string `json:"file"`
}

func (i Item) Title() string       { return i.File }
func (i Item) Description() string { return fmt.Sprintf("Print: %v", i.Print) }
func (i Item) FilterValue() string { return i.File }

type model struct {
	list     list.Model
	items    []Item
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case " ":
			index := m.list.Index()
			m.items[index].Print = !m.items[index].Print
			m.list.SetItems(itemsToListItems(m.items))
			saveItems(m.items)
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Bye!"
	}
	return docStyle.Render(m.list.View())
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func Main() {
	items := loadItems()

	m := model{
		list:  list.New(itemsToListItems(items), list.NewDefaultDelegate(), 0, 0),
		items: items,
	}
	m.list.Title = "JSONL Items"

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
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

func itemsToListItems(items []Item) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	return listItems
}
