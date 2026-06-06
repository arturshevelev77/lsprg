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

// PackageInfo хранит информацию об установленном пакете.
type PackageInfo struct {
	Name        string // Имя пакета
	Version     string // Версия
	Description string // Описание от разработчика
	Source      string // Источник (Core, Extra, CachyOS, AUR/Local и т.д.)
}

// model представляет состояние нашего Bubble Tea приложения.
type model struct {
	pattern    string        // Шаблон поиска (например, "*fetch" или "fetch")
	packages   []PackageInfo // Найденные пакеты
	err        error         // Ошибка в процессе работы (если есть)
	searching  bool          // Флаг активного поиска
}

// Сообщения для архитектуры Elm/Bubble Tea.
type searchResultMsg []PackageInfo
type errMsg error

// initialModel инициализирует модель на основе аргументов командной строки.
func initialModel(pattern string) model {
	return model{
		pattern:   pattern,
		searching: true,
	}
}

// Init запускает процесс сканирования пакетов при старте программы.
func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		pkgs, err := scanLocalPackages(m.pattern)
		if err != nil {
			return errMsg(err)
		}
		return searchResultMsg(pkgs)
	}
}

// Update обрабатывает сообщения и обновляет состояние.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultMsg:
		m.packages = msg
		m.searching = false
		return m, tea.Quit // Быстро выходим, так как нам нужен вывод в стиле "ls"

	case errMsg:
		m.err = msg
		m.searching = false
		return m, tea.Quit

	case tea.KeyMsg:
		// Прерывание по Ctrl+C
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

// View отвечает за красивый рендеринг в терминале с помощью lipgloss.
func (m model) View() string {
	// Инициализация стилей Lipgloss для "сочного" UI
	var (
		accentColor = lipgloss.Color("#FF5F87") // Розовый акцент
		cyanColor   = lipgloss.Color("#00F5D4") // Бирюзовый
		grayColor   = lipgloss.Color("#707070") // Серый для метаданных
		yellowColor = lipgloss.Color("#FFE15D") // Желтый для подсветки источников

		titleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E1E2E")).
				Background(accentColor).
				Padding(0, 1).
				Bold(true)

		patternStyle = lipgloss.NewStyle().
				Foreground(cyanColor).
				Bold(true)

		pkgNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

		descStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D9E0EE"))

		sourceStyle = lipgloss.NewStyle().
				Foreground(yellowColor).
				Italic(true)

		errStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)
	)

	if m.err != nil {
		return fmt.Sprintf("\n%s %v\n", errStyle.Render("Ошибка:"), m.err)
	}

	if m.searching {
		return fmt.Sprintf("\n⏳ Поиск пакетов по шаблону %s...\n", patternStyle.Render(m.pattern))
	}

	if len(m.packages) == 0 {
		return fmt.Sprintf("\n%s Нет совпадений по шаблону %s\n", 
			titleStyle.Render("lsprg"), 
			patternStyle.Render(m.pattern),
		)
	}

	// Строим результирующую таблицу вывода
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s Результаты для %s (найдено %d):\n\n", 
		titleStyle.Render("lsprg"), 
		patternStyle.Render(m.pattern), 
		len(m.packages),
	))

	// Отрендерим каждый найденный пакет компактно, в стиле "ls"
	for _, pkg := range m.packages {
		// Ограничиваем длину описания для компактности вывода
		shortDesc := pkg.Description
		if len(shortDesc) > 80 {
			shortDesc = shortDesc[:77] + "..."
		}

		// Форматируем строку: Имя [Версия] [Источник] - Описание
		pkgMeta := fmt.Sprintf("%s %s %s", 
			pkgNameStyle.Render(pkg.Name),
			lipgloss.NewStyle().Foreground(grayColor).Render(pkg.Version),
			sourceStyle.Render("["+pkg.Source+"]"),
		)

		// Добавляем отступы и красивую стрелочку
		b.WriteString(fmt.Sprintf("  ➔  %-50s %s\n", pkgMeta, descStyle.Render(shortDesc)))
	}
	b.WriteString("\n")

	return b.String()
}

// scanLocalPackages сканирует директорию /var/lib/pacman/local/ и извлекает информацию.
func scanLocalPackages(pattern string) ([]PackageInfo, error) {
	// Путь по умолчанию к локальной базе pacman
	dbPath := "/var/lib/pacman/local"
	
	// Проверим, существует ли директория. Если нет (например, запуск не на Arch Linux),
	// предоставим заглушку.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return getMockPackages(pattern), nil
	}

	files, err := os.ReadDir(dbPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать базу данных pacman: %w", err)
	}

	// Превращаем glob-шаблон в регулярное выражение
	regexPattern := globToRegex(pattern)
	re, err := regexp.Compile("(?i)" + regexPattern) // Регистронезависимый поиск
	if err != nil {
		return nil, fmt.Errorf("некорректный шаблон поиска: %w", err)
	}

	var results []PackageInfo

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dirName := file.Name()
		// Извлекаем ЧИСТОЕ имя пакета из названия директории (без версии)
		pkgName := splitNameVersion(dirName)

		// Теперь сопоставляем регулярное выражение именно с чистым именем пакета,
		// чтобы суффиксы версий вроде "-2.15.0-1" в названии папки не ломали поиск!
		if re.MatchString(pkgName) {
			descFilePath := filepath.Join(dbPath, dirName, "desc")
			pkg, err := parseDescFile(descFilePath)
			if err != nil {
				// Если не удалось распарсить конкретный desc-файл, создаем минимальную запись
				pkg = PackageInfo{
					Name:        pkgName,
					Version:     extractVersion(dirName, pkgName),
					Description: "Нет описания",
				}
			}
			
			if pkg.Name == "" {
				pkg.Name = pkgName
			}

			// Пытаемся определить источник (например, AUR, CachyOS, extra)
			pkg.Source = determineSource(dirName, pkg.Name)

			results = append(results, pkg)
		}
	}

	return results, nil
}

// parseDescFile парсит файл 'desc', вытаскивая имя, версию и описание.
func parseDescFile(path string) (PackageInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return PackageInfo{}, err
	}
	defer file.Close()

	var pkg PackageInfo
	scanner := bufio.NewScanner(file)
	
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Секции в pacman/desc начинаются с %
		if strings.HasPrefix(line, "%") && strings.HasSuffix(line, "%") {
			currentSection = line
			continue
		}

		switch currentSection {
		case "%NAME%":
			pkg.Name = line
		case "%VERSION%":
			pkg.Version = line
		case "%DESC%":
			// Описание может состоять из нескольких строк, собираем их
			if pkg.Description == "" {
				pkg.Description = line
			} else {
				pkg.Description += " " + line
			}
		}
	}

	return pkg, scanner.Err()
}

// globToRegex преобразует простой пользовательский фильтр (с символом *) в регулярку.
func globToRegex(pattern string) string {
	if pattern == "" || pattern == "*" {
		return ".*"
	}
	
	// Если пользователь не ввел звездочки, мы ищем вхождение подстроки (fetch -> .*fetch.*)
	if !strings.Contains(pattern, "*") {
		return ".*" + regexp.QuoteMeta(pattern) + ".*"
	}
	
	// Если звездочки есть, заменяем их на .* для гибкого поиска
	parts := strings.Split(pattern, "*")
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	return "^" + strings.Join(parts, ".*") + "$"
}

// splitNameVersion отделяет имя пакета от его версии на основе папки.
// Пример папки: "fastfetch-2.15.0-1" -> "fastfetch"
func splitNameVersion(dirName string) string {
	parts := strings.Split(dirName, "-")
	if len(parts) > 2 {
		// Обычно последние 2 элемента — это релиз и версия (например, "2.15.0" и "1")
		// Проверим, является ли предпоследний элемент началом версии (обычно начинается с цифры)
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 && (parts[i][0] >= '0' && parts[i][0] <= '9') {
				return strings.Join(parts[:i], "-")
			}
		}
		// Запасной вариант
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return dirName
}

// extractVersion вычленяет версию из названия директории, если не удалось прочитать desc-файл.
func extractVersion(dirName, pkgName string) string {
	version := strings.TrimPrefix(dirName, pkgName)
	return strings.TrimPrefix(version, "-")
}

// determineSource определяет репозиторий, из которого был установлен пакет.
func determineSource(dirName, pkgName string) string {
	lowerName := strings.ToLower(pkgName)
	lowerDir := strings.ToLower(dirName)
	
	// Эвристика для CachyOS
	if strings.Contains(lowerDir, "cachyos") || strings.Contains(lowerName, "cachyos") {
		return "cachyos"
	}
	
	// Эвристика для AUR пакетов
	if strings.HasSuffix(lowerName, "-git") || 
	   strings.HasSuffix(lowerName, "-bin") || 
	   strings.HasSuffix(lowerName, "-aur") || 
	   strings.HasSuffix(lowerName, "-yay") {
		return "AUR"
	}

	// Если в названии папки есть явное указание на репозиторий (иногда добавляется дистрибьюторами)
	// Иначе по умолчанию считаем системным pacman-пакетом
	return "pacman"
}

// getMockPackages создает демонстрационные данные для платформ без pacman.
func getMockPackages(pattern string) []PackageInfo {
	mocks := []PackageInfo{
		{Name: "neofetch", Version: "7.1.0-2", Description: "A fast, highly customizable system info script", Source: "pacman"},
		{Name: "fastfetch", Version: "2.15.0-1", Description: "Like neofetch, but much faster because written mostly in C", Source: "cachyos"},
		{Name: "hyfetch", Version: "1.4.11-1", Description: "Neofetch with pride flags! Light-weight system information tool", Source: "AUR"},
		{Name: "uwufetch", Version: "2.1.0-1", Description: "A meme system info tool for uvu boys and girls", Source: "AUR"},
		{Name: "cpufetch", Version: "1.0.5-1", Description: "Simple CPU architecture fetching tool", Source: "extra"},
		{Name: "onefetch", Version: "2.20.0-1", Description: "Git repository summary on your terminal", Source: "extra"},
	}

	regexPattern := globToRegex(pattern)
	re := regexp.MustCompile("(?i)" + regexPattern)

	var filtered []PackageInfo
	for _, pkg := range mocks {
		if re.MatchString(pkg.Name) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func main() {
	// По умолчанию ищем "fetch", чтобы при обычном запуске lsprg выдавал все фетчи на Arch/CachyOS
	pattern := "fetch"
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	p := tea.NewProgram(initialModel(pattern))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Критическая ошибка: %v\n", err)
		os.Exit(1)
	}
}
