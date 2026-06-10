# 📦 lsprg

A modern, fast, and beautiful CLI package search utility for Arch Linux. Built on a pure adrenaline vibe, powered by Go and the Charm CLI ecosystem.

---

## ✨ Features

* ⚡ **Blazing Fast:** Directly scans `/var/lib/pacman/local/` without spawning slow external `pacman` processes.
* 🎨 **Modern TUI:** Built using **Bubble Tea** and **Lipgloss** for that sleek, reactive terminal feel.
* 🔍 **Smart Search:** Supports glob patterns (like `*fetch*`) and automatically filters out version suffixes so your searches never break.
* 📦 **Repository Heuristics:** Distinguishes between standard packages, AUR, and CachyOS repos.

---

## 🛠️ Quick Start

You don't need a heavy setup. Build and run it with a single command:

```bash
go build -o lsprg main.go && ./lsprg

```

### Usage

```bash
./lsprg <search_pattern>
# Example: ./lsprg fetch

```

---

## ⚙️ Architecture: lsprg vs Old-school CLI

Most classic utilities (like `cbonsai`) are written in **C** using **ncurses** — a library from 1993.

`lsprg` takes a modern approach:

* **The Elm Architecture (TEA):** Predictable state management via Bubble Tea (`Init`, `Update`, `View`).
* **Asynchronous Engine:** Search tasks run safely in the background using `tea.Cmd` without freezing the terminal.
* **Declarative Styling:** Beautiful layouts and colors handled easily via Lipgloss.

---

## 🎸 Development Vibe & Philosophy

This project was hard-coded in a dark room under the heavy sounds of Siberian punk rock.

```text
Быть плоохим примером гараздо виселее...
Мама я люблю LINUX
Мама я ДРОЧУ на LINUX
Мама я пользуюсь VIMMMMM
МАМ-А Я ЛЮБЛ-Ю LINUX
А всё потому что я-я Крут

```

---

## 🤝 Contributed by

* **arturshevelev77** (*noob-in-coding* but coding on a pure vibe)


Закидывай в репозиторий! Как батя вернёт зарядник — расчехляй `vhs` или пиши экран, вставляй GIF-ку сразу под заголовком, и проект будет оформлен по всем канонам высшей лиги.

```
