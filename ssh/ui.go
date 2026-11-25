package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	accent = lipgloss.Color("#4f9cff")
	green  = lipgloss.Color("#3ddc97")
	muted  = lipgloss.Color("#7d8590")
	text   = lipgloss.Color("#e6edf3")

	titleSt  = lipgloss.NewStyle().Bold(true).Foreground(text)
	accentSt = lipgloss.NewStyle().Foreground(accent)
	mutedSt  = lipgloss.NewStyle().Foreground(muted)
	greenSt  = lipgloss.NewStyle().Foreground(green)
	h2St     = lipgloss.NewStyle().Bold(true).Foreground(accent).MarginTop(1)
	selSt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#0b0e14")).Background(accent)
)

type tab int

const (
	tabProjects tab = iota
	tabAbout
	tabContact
)

var tabNames = []string{"Projects", "About", "Contact"}

type model struct {
	site    *Site
	w, h    int
	tab     tab
	cursor  int
	reading bool // true while a case study is open
	vp      viewport.Model
}

func newModel(s *Site, w, h int) model {
	if w < 20 {
		w = 80
	}
	if h < 8 {
		h = 24
	}
	m := model{site: s, w: w, h: h}
	m.vp = viewport.New(w, h-4)
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width < 20 || msg.Height < 8 {
			return m, nil
		}
		m.w, m.h = msg.Width, msg.Height
		m.vp.Width, m.vp.Height = m.w, m.h-4
		if m.reading {
			m.vp.SetContent(m.caseStudy(m.site.Projects[m.cursor]))
		}
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" || k == "q" && !m.reading {
			return m, tea.Quit
		}
		if m.reading {
			switch k {
			case "esc", "backspace", "left", "h", "q":
				m.reading = false
				return m, nil
			}
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
		switch k {
		case "tab", "right", "l":
			m.tab = (m.tab + 1) % 3
		case "shift+tab", "left", "h":
			m.tab = (m.tab + 2) % 3
		case "1":
			m.tab = tabProjects
		case "2":
			m.tab = tabAbout
		case "3":
			m.tab = tabContact
		case "up", "k":
			if m.tab == tabProjects && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.tab == tabProjects && m.cursor < len(m.site.Projects)-1 {
				m.cursor++
			}
		case "enter":
			if m.tab == tabProjects {
				m.reading = true
				m.vp.Width, m.vp.Height = m.w, m.h-4
				m.vp.SetContent(m.caseStudy(m.site.Projects[m.cursor]))
				m.vp.GotoTop()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	body := ""
	switch {
	case m.reading:
		body = m.vp.View()
	case m.tab == tabProjects:
		body = m.projectList()
	case m.tab == tabAbout:
		body = m.about()
	default:
		body = m.contact()
	}
	// Pad/clip the body so the footer stays on the last line.
	lines := strings.Split(body, "\n")
	room := m.h - 4
	if room < 1 {
		room = 1
	}
	if len(lines) > room {
		lines = lines[:room]
	}
	for len(lines) < room {
		lines = append(lines, "")
	}
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m model) header() string {
	var tabs []string
	for i, n := range tabNames {
		label := fmt.Sprintf(" %d %s ", i+1, n)
		if tab(i) == m.tab && !m.reading {
			tabs = append(tabs, selSt.Render(label))
		} else {
			tabs = append(tabs, mutedSt.Render(label))
		}
	}
	name := titleSt.Render("Pratham Bhavsar") + mutedSt.Render("  backend & full-stack engineer")
	return name + "\n" + strings.Join(tabs, " ") + "\n" + mutedSt.Render(strings.Repeat("─", m.w))
}

func (m model) footer() string {
	help := "↑↓ move · enter open · tab switch · q quit"
	if m.reading {
		help = "↑↓ scroll · esc back"
	}
	return mutedSt.Render(help)
}

func (m model) projectList() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, p := range m.site.Projects {
		cur := "  "
		title := titleSt.Render(p.Title)
		if i == m.cursor {
			cur = accentSt.Render("▌ ")
			title = accentSt.Bold(true).Render(p.Title)
		}
		b.WriteString(cur + title + "  " + mutedSt.Render(p.Role) + "\n")
		if i == m.cursor {
			w := m.w - 4
			b.WriteString("  " + wrap(p.Line, w, "  ") + "\n")
			b.WriteString("  " + mutedSt.Render(strings.Join(p.Spec, " · ")) + "\n")
			b.WriteString("  " + greenSt.Render(strings.Join(p.Stack, "  ")) + "\n")
		}
	}
	return b.String()
}

func (m model) caseStudy(p Project) string {
	w := m.w - 2
	if w > 90 {
		w = 90
	}
	var b strings.Builder
	b.WriteString("\n" + titleSt.Render(p.Title) + "\n")
	b.WriteString(wrap(p.Line, w, "") + "\n")
	b.WriteString(mutedSt.Render(strings.Join(p.Spec, " · ")) + "\n")
	b.WriteString(greenSt.Render(strings.Join(p.Stack, "  ")) + "\n")
	for _, l := range p.Links {
		b.WriteString(accentSt.Render(l.Href) + mutedSt.Render("  ("+l.Label+")") + "\n")
	}
	if len(p.Links) == 0 {
		b.WriteString(mutedSt.Render("private, no public link") + "\n")
	}
	for _, block := range strings.Split(strings.TrimSpace(p.Content), "\n\n") {
		lines := strings.Split(block, "\n")
		switch {
		case strings.HasPrefix(block, "## "):
			b.WriteString(h2St.Render(strip(block[3:])) + "\n")
		case strings.HasPrefix(lines[0], "- "):
			for _, l := range lines {
				b.WriteString(accentSt.Render("• ") + wrap(strip(strings.TrimPrefix(l, "- ")), w-2, "  ") + "\n")
			}
		default:
			b.WriteString("\n" + wrap(strip(strings.Join(lines, " ")), w, "") + "\n")
		}
	}
	return b.String()
}

func (m model) about() string {
	w := m.w - 4
	if w > 84 {
		w = 84
	}
	t := "I started out writing Flutter apps and moved to backends because that's where the interesting problems were. These days I mostly write Rust (axum, SQLx) and Go, with Postgres underneath and React or Next.js on top. Lately I've been building open-source tools I'd want to use myself: an affiliate engine, a self-hosted support inbox, a study tool, and a terminal workspace for running coding agents."
	roles := [][2]string{
		{"Pixel Gen", "Software Development Engineer · 08/2025 – now"},
		{"Nyxos", "Developer, client projects · now · nyxos.in"},
		{"Prudent Corporate", "Flutter Developer · 09/2025 – 02/2026"},
		{"Nautilus Techlabs", "Flutter Developer · 02/2024 – 09/2025"},
		{"Exito Systems", "Flutter Intern · 11/2023 – 02/2024"},
	}
	var b strings.Builder
	b.WriteString("\n" + wrap(t, w, "") + "\n\n")
	b.WriteString(h2St.Render("Experience") + "\n")
	for _, r := range roles {
		b.WriteString(titleSt.Render(r[0]) + "  " + mutedSt.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + mutedSt.Render("B.Tech Computer Engineering, Gandhinagar Institute of Technology, 2024"))
	return b.String()
}

func (m model) contact() string {
	return "\n" +
		titleSt.Render("Email     ") + accentSt.Render("pratham.bhavsar501@gmail.com") + "\n" +
		titleSt.Render("GitHub    ") + accentSt.Render("github.com/PrathamBhavsar") + "\n" +
		titleSt.Render("Web       ") + accentSt.Render("prathambhavsar.github.io") + "\n" +
		titleSt.Render("Location  ") + "Ahmedabad, India\n"
}

// wrap word-wraps s to width, indenting continuation lines.
func wrap(s string, width int, indent string) string {
	if width < 10 {
		width = 10
	}
	var lines []string
	line := ""
	for _, w := range strings.Fields(s) {
		if line != "" && lipgloss.Width(line)+1+lipgloss.Width(w) > width {
			lines = append(lines, line)
			line = w
			continue
		}
		if line == "" {
			line = w
		} else {
			line += " " + w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"+indent)
}
