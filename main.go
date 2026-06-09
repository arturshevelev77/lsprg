package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// инфа о пакете
type pkg struct {
	name string
	ver  string
	desc string
	src  string
}

type model struct {
	pat  string
	pkgs []pkg
	err  error
	load bool
}

func main() {
	pat := "fetch"
	if len(os.Args) > 1 {
		pat = os.Args[1]
	}
	// запускаем прогу
	p := tea.NewProgram(model{pat: pat, load: true})
	if _, err := p.Run(); err != nil {
		fmt.Printf("ошибка %v\n", err)
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		// ищем пакеты
		d := "/var/lib/pacman/local" // расположение пакетиков
		var list []pkg
		files, err := os.ReadDir(d)
		if err != nil {
			return nil
		}
		re := regexp.MustCompile("(?i)" + strings.ReplaceAll(m.pat, "*", ".*"))
		for _, f := range files {
			if f.IsDir() && re.MatchString(f.Name()) {
				// Пытаемся сука вывести инфу у этих молчащих портизанов где блять мой aur и pacman
				p := pkg{name: f.Name()}
				list = append(list, p)
			}
		}
		return list
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []pkg:
		m.pkgs = msg
		m.load = false
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.load {
		return "ищем..."
	}
	// Отображаем наш обдолбаный список пакетов
	s := "найдено " + fmt.Sprintf("%d", len(m.pkgs)) + "\n"
	for _, p := range m.pkgs {
		i := "◎"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("   ", i, p.name) + "\n"
	}
	return s
}
