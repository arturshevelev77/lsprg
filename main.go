package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	CorePoint string
}

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
	cfg  Config 
}

func main() {
	pat := "fetch"
	if len(os.Args) > 1 {
		pat = os.Args[1]
	}

	config := loadOrInitConfig()

	p := tea.NewProgram(model{pat: pat, load: true, cfg: config})
	if _, err := p.Run(); err != nil {
		fmt.Printf("error %v\n", err)
	}
}

func loadOrInitConfig() Config {
	homeDir, _ := os.UserHomeDir()
	dirPath := filepath.Join(homeDir, ".config", "lsprg")
	filePath := filepath.Join(dirPath, "lsprg.conf")

	cfg := Config{CorePoint: "◎"}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0755)
		os.WriteFile(filePath, []byte("core_point=◎"), 0644)
		return cfg
	}

	file, err := os.Open(filePath)
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "core_point=") {
			cfg.CorePoint = strings.TrimPrefix(line, "core_point=")
		}
	}
	return cfg
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		d := "/var/lib/pacman/local"
		var list []pkg
		files, err := os.ReadDir(d)
		if err != nil {
			return nil 
		}
		
		re := regexp.MustCompile("(?i)" + strings.ReplaceAll(m.pat, "*", ".*"))
		
		for _, f := range files {
			if !f.IsDir() || f.Name() == "ALPM_DB_VERSION" {
				continue 
			}
			
			if re.MatchString(f.Name()) {
				// По дефолту пишем aur, если в desc ничего другого не найдем
				repo := "aur" 

				// Быстро смотрим в desc пакета
				descFile, err := os.Open(filepath.Join(d, f.Name(), "desc"))
				if err == nil {
					scanner := bufio.NewScanner(descFile)
					isDb := false
					for scanner.Scan() {
						line := scanner.Text()
						if line == "%DB%" {
							isDb = true
							continue
						}
						if isDb {
							repo = line // нашли репозиторий (core, extra и т.д.)
							break
						}
					}
					descFile.Close()
				}

				p := pkg{name: f.Name(), src: repo}
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
	}
	return m, nil
}

func (m model) View() string {
	if m.load {
		return "search..."
	}
	
	s := "[lsprg] finding " + fmt.Sprint(len(m.pkgs)) + "\n"
	for _, p := range m.pkgs {
		// Поправил тут p.name и добавил вывод p.src, как у тебя и было задумано
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("    "+m.cfg.CorePoint+" ", p.name)
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("178")).Render(" (") + p.src + ")" + "\n"
	}
	return s
}
