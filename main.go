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

// Config хранит наши параметры
type Config struct {
	CorePoint string
}

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
	cfg  Config // теперь конфиг живет тут
}

func main() {
	pat := "fetch"
	if len(os.Args) > 1 {
		pat = os.Args[1]
	}

	// Инициализируем конфиг перед запуском программы
	config := loadOrInitConfig()

	// запускаем прогу
	p := tea.NewProgram(model{pat: pat, load: true, cfg: config})
	if _, err := p.Run(); err != nil {
		fmt.Printf("ошибка %v\n", err)
	}
}

// loadOrInitConfig проверяет наличие конфига, если нет — создает дефолтный
func loadOrInitConfig() Config {
	homeDir, _ := os.UserHomeDir()
	dirPath := filepath.Join(homeDir, ".config", "lsprg")
	filePath := filepath.Join(dirPath, "lsprg.conf")

	// Дефолтное значение
	cfg := Config{CorePoint: "◎"}

	// Если файла нет, создаем его с дефолтом
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0755)
		os.WriteFile(filePath, []byte("core_point=◎"), 0644)
		return cfg
	}

	// Если файл есть, читаем его
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
            return nil // Здесь лучше возвращать структуру с ошибкой
        }
        
        re := regexp.MustCompile("(?i)" + strings.ReplaceAll(m.pat, "*", ".*"))
        
        for _, f := range files {
            // ДОБАВЛЯЕМ ПРОВЕРКУ:
            // 1. Должно быть директорией
            // 2. Имя не должно быть "ALPM_DB_VERSION"
            if !f.IsDir() || f.Name() == "ALPM_DB_VERSION" {
                continue 
            }
            
            if re.MatchString(f.Name()) {
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
	}
	return m, nil
}

func (m model) View() string {
	if m.load {
		return "ищем..."
	}
	// Отображаем наш обдолбаный список пакетов
	s := "найдено " + fmt.Sprint(len(m.pkgs)) + "\n"
	for _, p := range m.pkgs {
		// используем символ из нашего загруженного конфига
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("   "+m.cfg.CorePoint, p.name) + "\n"
	}
	return s
}
