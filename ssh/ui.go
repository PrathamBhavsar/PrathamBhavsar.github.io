package main

// Look and feel follows the "Mono" theme from my Kestrel TUI: near-black,
// hairline panels with inset titles, one blue accent, inverse-video selection,
// chips for state, a top bar with tabs and a status line with key hints.

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

const (
	cBG     = lipgloss.Color("#0a0a0a")
	cFG     = lipgloss.Color("#ededed")
	cDim    = lipgloss.Color("#8a8a8a")
	cMute   = lipgloss.Color("#555555")
	cBorder = lipgloss.Color("#2a2a2a")
	cAccent = lipgloss.Color("#3291ff")
	cTeal   = lipgloss.Color("#50e3c2")
	cOK     = lipgloss.Color("#00c853")
)

// theme holds every style, built from the per-session renderer so colours follow
// the visitor's terminal instead of the server's (which has no TTY and would
// render everything as plain text).
type theme struct {
	fg, bold, dim, mute, accent, teal, ok, sel lipgloss.Style
	border                                     lipgloss.Color
	base                                       string // SGR for the page background + default foreground
}

func newTheme(r *lipgloss.Renderer) theme {
	st := func(c lipgloss.Color) lipgloss.Style { return r.NewStyle().Foreground(c) }
	p := r.ColorProfile()
	base := ""
	if p != termenv.Ascii {
		base = termenv.CSI + p.Color(string(cBG)).Sequence(true) + "m" + termenv.CSI + p.Color(string(cFG)).Sequence(false) + "m"
	}
	return theme{
		fg: st(cFG), bold: st(cFG).Bold(true), dim: st(cDim), mute: st(cMute),
		accent: st(cAccent), teal: st(cTeal), ok: st(cOK),
		sel:    r.NewStyle().Foreground(cBG).Background(cFG).Bold(true),
		border: cBorder, base: base,
	}
}

func (t theme) chip(text string, bg lipgloss.Color) string {
	return t.fg.UnsetForeground().Foreground(cBG).Background(bg).Bold(true).Render(" " + text + " ")
}

// link renders text as an OSC 8 hyperlink (terminals without support show plain text).
func (t theme) link(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + t.accent.Underline(true).Render(text) + "\x1b]8;;\x1b\\"
}

// paint gives the whole screen the dark background: every SGR reset is followed by
// the base colours again, and each row is padded to the full width.
func (t theme) paint(view string, w int) string {
	if t.base == "" {
		return view
	}
	lines := strings.Split(view, "\n")
	for i, l := range lines {
		l = strings.ReplaceAll(l, "\x1b[0m", "\x1b[0m"+t.base)
		l = strings.ReplaceAll(l, "\x1b[m", "\x1b[m"+t.base)
		pad := max(0, w-lipgloss.Width(l))
		lines[i] = t.base + l + strings.Repeat(" ", pad) + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}

func short(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://"), "www.")
}

type tab int

const (
	tabProjects tab = iota
	tabAbout
	tabLinks
)

var tabNames = []string{"Projects", "About", "Links"}

type pane int

const (
	paneList pane = iota
	paneBody
)

type model struct {
	t         theme
	site      *Site
	w, h      int
	tab       tab
	focus     pane
	cursor    int
	scroll    map[string]int // per-view scroll offset
	filter    string
	filtering bool
	help      bool
}

func newModel(s *Site, r *lipgloss.Renderer, w, h int) model {
	if w < 20 {
		w = 100
	}
	if h < 8 {
		h = 30
	}
	return model{t: newTheme(r), site: s, w: w, h: h, scroll: map[string]int{}}
}

func (m model) Init() tea.Cmd { return nil }

// visible returns the project indexes that match the current filter.
func (m model) visible() []int {
	var out []int
	f := strings.ToLower(m.filter)
	for i, p := range m.site.Projects {
		hay := strings.ToLower(p.Title + " " + p.Line + " " + strings.Join(p.Stack, " ") + " " + p.Role)
		if f == "" || strings.Contains(hay, f) {
			out = append(out, i)
		}
	}
	return out
}

func (m model) clampedCursor() int {
	v := m.visible()
	c := m.cursor
	if c >= len(v) {
		c = len(v) - 1
	}
	if c < 0 {
		c = 0
	}
	return c
}

func (m model) current() (Project, bool) {
	v := m.visible()
	if len(v) == 0 {
		return Project{}, false
	}
	return m.site.Projects[v[m.clampedCursor()]], true
}

func (m model) viewKey() string {
	if m.tab == tabProjects {
		if p, ok := m.current(); ok {
			return "p:" + p.Slug
		}
	}
	return fmt.Sprintf("t:%d", m.tab)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width >= 20 && msg.Height >= 8 {
			m.w, m.h = msg.Width, msg.Height
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			k := m.viewKey()
			switch msg.Button {
			case tea.MouseButtonWheelDown:
				m.scroll[k] += 3
			case tea.MouseButtonWheelUp:
				m.scroll[k] = max(0, m.scroll[k]-3)
			}
		}
	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	if m.filtering {
		switch k {
		case "esc":
			m.filtering, m.filter, m.cursor = false, "", 0
		case "enter":
			m.filtering = false
		case "backspace":
			if r := []rune(m.filter); len(r) > 0 {
				m.filter = string(r[:len(r)-1])
			}
			m.cursor = 0
		default:
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m.cursor = 0
			}
		}
		return m, nil
	}
	if m.help {
		m.help = false
		return m, nil
	}
	vk := m.viewKey()
	switch k {
	case "q":
		return m, tea.Quit
	case "?":
		m.help = true
	case "1":
		m.tab, m.focus = tabProjects, paneList
	case "2":
		m.tab = tabAbout
	case "3":
		m.tab = tabLinks
	case "]":
		m.tab = (m.tab + 1) % 3
	case "[":
		m.tab = (m.tab + 2) % 3
	case "tab":
		if m.tab == tabProjects {
			m.focus = 1 - m.focus
		} else {
			m.tab = (m.tab + 1) % 3
		}
	case "/":
		m.tab, m.filtering, m.focus = tabProjects, true, paneList
	case "esc":
		if m.filter != "" {
			m.filter, m.cursor = "", 0
		} else if m.focus == paneBody && m.tab == tabProjects {
			m.focus = paneList
		}
	case "enter", "right", "l":
		if m.tab == tabProjects && m.focus == paneList {
			m.focus = paneBody
		}
	case "left", "h", "backspace":
		if m.tab == tabProjects && m.focus == paneBody {
			m.focus = paneList
		}
	case "up", "k":
		if m.tab == tabProjects && m.focus == paneList {
			m.cursor = max(0, m.clampedCursor()-1)
		} else {
			m.scroll[vk] = max(0, m.scroll[vk]-1)
		}
	case "down", "j":
		if m.tab == tabProjects && m.focus == paneList {
			m.cursor = min(len(m.visible())-1, m.clampedCursor()+1)
		} else {
			m.scroll[vk]++
		}
	case "pgdown", " ", "f":
		m.scroll[vk] += m.h - 6
	case "pgup", "b":
		m.scroll[vk] = max(0, m.scroll[vk]-(m.h-6))
	case "g", "home":
		m.scroll[vk] = 0
	case "G", "end":
		m.scroll[vk] = 1 << 20 // clamped when rendered
	}
	return m, nil
}

/* ---------- layout ---------- */

func (m model) View() string {
	body := m.h - 2
	var main string
	switch {
	case m.help:
		main = m.helpPanel(m.w, body)
	case m.tab == tabProjects:
		main = m.projectsView(m.w, body)
	case m.tab == tabAbout:
		main = m.aboutView(m.w, body)
	default:
		main = m.linksView(m.w, body)
	}
	return m.t.paint(m.topBar(m.w)+"\n"+main+"\n"+m.statusLine(m.w), m.w)
}

func (m model) topBar(w int) string {
	left := m.t.chip("pb", cAccent) + " "
	for i, n := range tabNames {
		label := fmt.Sprintf(" %d %s ", i+1, n)
		if tab(i) == m.tab {
			left += m.t.sel.Render(label)
		} else {
			left += m.t.dim.Render(label)
		}
	}
	right := m.t.bold.Render(m.site.Profile.Name) + m.t.dim.Render("  "+m.site.Profile.Location+" ") + m.t.ok.Render("●") + m.t.dim.Render(" open to work")
	return spread(left, right, w)
}

func (m model) statusLine(w int) string {
	mode := "BROWSE"
	hints := [][2]string{{"↑↓", "move"}, {"⏎", "open"}, {"⇥", "focus"}, {"/", "filter"}, {"1-3", "tabs"}, {"?", "help"}, {"q", "quit"}}
	switch {
	case m.filtering:
		mode = "FILTER"
		hints = [][2]string{{"type", "to filter"}, {"⏎", "keep"}, {"esc", "clear"}}
	case m.tab == tabProjects && m.focus == paneBody:
		mode = "READ"
		hints = [][2]string{{"↑↓", "scroll"}, {"space", "page"}, {"g/G", "ends"}, {"esc", "back"}, {"q", "quit"}}
	case m.tab != tabProjects:
		hints = [][2]string{{"↑↓", "scroll"}, {"[ ]", "tabs"}, {"?", "help"}, {"q", "quit"}}
	}
	left := m.t.chip(mode, cAccent) + " "
	for _, h := range hints {
		left += m.t.bold.Render(h[0]) + m.t.dim.Render(" "+h[1]) + "   "
	}
	return spread(left, m.t.dim.Render("ssh ssh.prathambhavsar.in"), w)
}

func spread(left, right string, w int) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return ansi.Truncate(left, w, "")
	}
	return left + strings.Repeat(" ", gap) + right
}

// panel draws a hairline box with an inset title and optional right-aligned meta.
func (m model) panel(title, meta string, lines []string, w, h int, focused bool) string {
	if w < 8 || h < 3 {
		return ""
	}
	bc, tc := cBorder, m.t.dim
	if focused {
		bc, tc = cAccent, m.t.bold
	}
	b := m.t.fg.Foreground(bc)
	inner := w - 2
	t := " " + title + " "
	mt := ""
	if meta != "" {
		mt = " " + meta + " "
	}
	fill := max(0, inner-1-lipgloss.Width(t)-lipgloss.Width(mt)-1)
	top := b.Render("┌─") + tc.Render(t) + b.Render(strings.Repeat("─", fill)) + m.t.dim.Render(mt) + b.Render("─┐")
	out := []string{top}
	for i := 0; i < h-2; i++ {
		l := ""
		if i < len(lines) {
			l = lines[i]
		}
		l = ansi.Truncate(l, inner-2, "…")
		pad := max(0, inner-2-lipgloss.Width(l))
		out = append(out, b.Render("│")+" "+l+strings.Repeat(" ", pad)+" "+b.Render("│"))
	}
	out = append(out, b.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(out, "\n")
}

// window slices wrapped lines for scrolling and reports the scroll percentage.
func (m model) window(lines []string, key string, rows int) ([]string, string) {
	mx := max(0, len(lines)-rows)
	off := min(max(m.scroll[key], 0), mx)
	m.scroll[key] = off
	end := min(off+rows, len(lines))
	meta := ""
	if mx > 0 {
		meta = fmt.Sprintf("%d%%", off*100/mx)
	}
	return lines[off:end], meta
}

/* ---------- projects ---------- */

func (m model) projectsView(w, h int) string {
	// <80: one pane at a time. 80-119: list + reader. 120+: list + reader + details.
	if w < 80 {
		if m.focus == paneList {
			return m.listPanel(w, h)
		}
		return m.readerPanel(w, h)
	}
	if w < 120 {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.listPanel(30, h), m.readerPanel(w-30, h))
	}
	listW, sideW := 34, 40
	return lipgloss.JoinHorizontal(lipgloss.Top, m.listPanel(listW, h), m.readerPanel(w-listW-sideW, h), m.detailPanel(sideW, h))
}

func (m model) roleChip(role string) string {
	if role == "client" {
		return m.t.chip("CLIENT", cTeal)
	}
	return m.t.chip("OWN", cAccent)
}

func (m model) listPanel(w, h int) string {
	var lines []string
	if m.filtering || m.filter != "" {
		cur := ""
		if m.filtering {
			cur = m.t.accent.Render("█")
		}
		lines = append(lines, m.t.accent.Render("/ ")+m.t.fg.Render(m.filter)+cur, "")
	}
	v := m.visible()
	cursor := m.clampedCursor()
	inner := w - 4
	for i, idx := range v {
		p := m.site.Projects[idx]
		name := " " + p.Title
		tag := p.Role
		gap := strings.Repeat(" ", max(1, inner-lipgloss.Width(name)-lipgloss.Width(tag)-1))
		if i == cursor {
			lines = append(lines, m.t.sel.Render(ansi.Truncate(name+gap+tag+" ", inner, "")))
			lines = append(lines, m.t.dim.Render(" "+ansi.Truncate(strings.Join(p.Stack[:min(3, len(p.Stack))], " · "), inner-1, "…")))
		} else {
			lines = append(lines, m.t.bold.Render(name)+m.t.mute.Render(gap+tag))
		}
	}
	if len(v) == 0 {
		lines = append(lines, m.t.dim.Render(" no matches"))
	}
	return m.panel("Projects", fmt.Sprintf("%d", len(v)), lines, w, h, m.focus == paneList)
}

func (m model) readerPanel(w, h int) string {
	p, ok := m.current()
	if !ok {
		return m.panel("Case study", "", []string{m.t.dim.Render("nothing selected")}, w, h, m.focus == paneBody)
	}
	cw := min(w-6, 92)
	lines := []string{m.t.bold.Render(p.Title) + "  " + m.roleChip(p.Role), ""}
	for _, x := range strings.Split(wrap(p.Line, cw), "\n") {
		lines = append(lines, m.t.fg.Render(x))
	}
	if p.Note != "" {
		lines = append(lines, "")
		for _, x := range strings.Split(wrap(p.Note, cw), "\n") {
			lines = append(lines, m.t.dim.Render(x))
		}
	}
	for _, block := range strings.Split(strings.TrimSpace(p.Content), "\n\n") {
		bl := strings.Split(block, "\n")
		switch {
		case strings.HasPrefix(block, "## "):
			lines = append(lines, "", m.t.accent.Bold(true).Render(strip(block[3:])))
		case strings.HasPrefix(bl[0], "- "):
			lines = append(lines, "")
			for _, l := range bl {
				for i, x := range strings.Split(wrap(strip(strings.TrimPrefix(l, "- ")), cw-2), "\n") {
					pre := "  "
					if i == 0 {
						pre = m.t.accent.Render("• ")
					}
					lines = append(lines, pre+m.t.fg.Render(x))
				}
			}
		default:
			lines = append(lines, "")
			for _, x := range strings.Split(wrap(strip(strings.Join(bl, " ")), cw), "\n") {
				lines = append(lines, m.t.fg.Render(x))
			}
		}
	}
	vis, meta := m.window(lines, "p:"+p.Slug, h-2)
	return m.panel(p.Title, meta, vis, w, h, m.focus == paneBody)
}

func (m model) detailPanel(w, h int) string {
	p, ok := m.current()
	if !ok {
		return m.panel("Details", "", nil, w, h, false)
	}
	l := []string{m.t.dim.Render("scope")}
	for _, s := range p.Spec {
		l = append(l, m.t.fg.Render(s))
	}
	l = append(l, "", m.t.dim.Render("stack"))
	for _, s := range p.Stack {
		l = append(l, m.t.teal.Render(s))
	}
	l = append(l, "", m.t.dim.Render("links"))
	if len(p.Links) == 0 {
		l = append(l, m.t.mute.Render("private, no public link"))
	}
	for _, k := range p.Links {
		l = append(l, m.t.chip(strings.ToUpper(k.Kind), cOK), m.t.link(k.Href, short(k.Href)))
	}
	return m.panel("Details", "", l, w, h, false)
}

/* ---------- about / links / help ---------- */

func (m model) aboutView(w, h int) string {
	if w < 100 {
		return m.aboutBody(w, h)
	}
	pr := m.site.Profile
	l := []string{m.t.bold.Render(pr.Name), m.t.dim.Render(pr.Tagline), m.t.dim.Render(pr.Location), "",
		m.t.dim.Render("email"), m.t.link("mailto:"+pr.Email, pr.Email), ""}
	for _, k := range pr.Links {
		l = append(l, m.t.dim.Render(strings.ToLower(k.Label)), m.t.link(k.URL, short(k.URL)), "")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, m.panel("Profile", "", l, 38, h, false), m.aboutBody(w-38, h))
}

func (m model) aboutBody(w, h int) string {
	pr := m.site.Profile
	cw := min(w-6, 92)
	var l []string
	for _, p := range pr.About {
		for _, x := range strings.Split(wrap(p, cw), "\n") {
			l = append(l, m.t.fg.Render(x))
		}
		l = append(l, "")
	}
	l = append(l, m.t.accent.Bold(true).Render("Skills"), "")
	for _, x := range strings.Split(wrap(strings.Join(pr.Skills, "  ·  "), cw), "\n") {
		l = append(l, m.t.teal.Render(x))
	}
	l = append(l, "", m.t.accent.Bold(true).Render("Experience"))
	for _, e := range pr.Experience {
		org := m.t.bold.Render(e.Org)
		if e.URL != "" {
			org = m.t.link(e.URL, e.Org)
		}
		l = append(l, "", org+m.t.dim.Render("  "+e.Role+"  ·  "+e.When))
		for _, x := range strings.Split(wrap(e.Desc, cw), "\n") {
			l = append(l, m.t.fg.Render(x))
		}
	}
	l = append(l, "", m.t.accent.Bold(true).Render("Education"), "")
	for _, e := range pr.Education {
		l = append(l, m.t.bold.Render(e.Org)+m.t.dim.Render("  "+e.Role+"  ·  "+e.When))
	}
	vis, meta := m.window(l, "t:1", h-2)
	return m.panel("About", meta, vis, w, h, true)
}

func (m model) linksView(w, h int) string {
	pr := m.site.Profile
	var l []string
	row := func(label, text, url string) {
		l = append(l, m.t.dim.Render(fmt.Sprintf("%-14s", label))+m.t.link(url, text))
	}
	row("email", pr.Email, "mailto:"+pr.Email)
	for _, k := range pr.Links {
		row(strings.ToLower(k.Label), k.URL, k.URL)
	}
	l = append(l, m.t.dim.Render(fmt.Sprintf("%-14s", "location"))+m.t.fg.Render(pr.Location))
	l = append(l, "", m.t.accent.Bold(true).Render("Projects"), "")
	for _, p := range m.site.Projects {
		if len(p.Links) == 0 {
			l = append(l, m.t.dim.Render(fmt.Sprintf("%-14s", strings.ToLower(p.Title)))+m.t.mute.Render("private"))
			continue
		}
		for i, k := range p.Links {
			name := ""
			if i == 0 {
				name = strings.ToLower(p.Title)
			}
			l = append(l, m.t.dim.Render(fmt.Sprintf("%-14s", name))+m.t.link(k.Href, k.Href))
		}
	}
	l = append(l, "", m.t.mute.Render("Links are clickable in terminals that support it (iTerm2, kitty, WezTerm, Windows Terminal)."))
	vis, meta := m.window(l, "t:2", h-2)
	return m.panel("Links", meta, vis, w, h, true)
}

func (m model) helpPanel(w, h int) string {
	rows := [][2]string{
		{"↑ ↓  j k", "move / scroll"},
		{"⏎  →", "open the selected project"},
		{"⇥", "switch between list and reader"},
		{"/", "filter projects by name, stack or role"},
		{"1 2 3  [ ]", "Projects, About, Links"},
		{"space b g G", "page down, page up, top, bottom"},
		{"mouse wheel", "scroll"},
		{"esc", "back / clear filter"},
		{"q  ctrl+c", "quit"},
	}
	var l []string
	for _, r := range rows {
		l = append(l, m.t.bold.Render(fmt.Sprintf("%-16s", r[0]))+m.t.dim.Render(r[1]))
	}
	l = append(l, "", m.t.mute.Render("press any key to close"))
	return m.panel("Help", "", l, w, h, true)
}

/* ---------- text helpers ---------- */

// wrap word-wraps plain text to width.
func wrap(s string, width int) string {
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
	return strings.Join(lines, "\n")
}
