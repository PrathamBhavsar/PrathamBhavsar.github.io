// portfolio-ssh serves the portfolio as a terminal app over SSH.
//
//	ssh pratham.example        (no password, no key, no shell)
//
// Every connection gets the TUI and nothing else: no shell, no exec, no
// port forwarding, no file transfer.
package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/muesli/termenv"
	"golang.org/x/time/rate"
)

func main() {
	addr := flag.String("addr", ":22", "listen address")
	hostKey := flag.String("host-key", ".ssh/id_ed25519", "host key path (created on first run, keep it stable)")
	idle := flag.Duration("idle", 10*time.Minute, "disconnect idle sessions after this long")
	maxLife := flag.Duration("max", 30*time.Minute, "hard cap on one session")
	flag.Parse()

	site, err := loadSite()
	if err != nil {
		log.Fatal("load content", "err", err)
	}

	// Per IP: one new connection every 2s, burst of 5. Plenty for a human.
	limiter := ratelimiter.NewRateLimiter(rate.Every(2*time.Second), 5, 4096)

	srv, err := wish.NewServer(
		wish.WithAddress(*addr),
		wish.WithHostKeyPath(*hostKey),
		wish.WithIdleTimeout(*idle),
		wish.WithMaxTimeout(*maxLife),
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				pty, _, _ := s.Pty()
				return newModel(site, sessionRenderer(s), pty.Window.Width, pty.Window.Height), []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
			}),
			activeterm.Middleware(), // a PTY is required; `ssh host cmd` is refused
			ratelimiter.Middleware(limiter),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("server", "err", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Info("listening", "addr", *addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			log.Fatal("serve", "err", err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// sessionEnv exposes the SSH client's environment (plus its PTY's TERM) to termenv.
type sessionEnv struct{ kv map[string]string }

func (e sessionEnv) Environ() []string {
	out := make([]string, 0, len(e.kv))
	for k, v := range e.kv {
		out = append(out, k+"="+v)
	}
	return out
}
func (e sessionEnv) Getenv(k string) string { return e.kv[k] }

// sessionRenderer builds a colour renderer for this visitor's terminal.
// It deliberately avoids wish's MakeRenderer, which asks the client for its
// background colour and waits for an answer. Many SSH clients don't forward
// COLORTERM, so a known TERM is taken as 256 colours (the safe floor and cap)
// and truecolor is used only when the client says it can.
func sessionRenderer(s ssh.Session) *lipgloss.Renderer {
	env := map[string]string{}
	for _, kv := range s.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	if pty, _, ok := s.Pty(); ok && env["TERM"] == "" {
		env["TERM"] = pty.Term
	}
	r := lipgloss.NewRenderer(s, termenv.WithEnvironment(sessionEnv{env}), termenv.WithUnsafe(), termenv.WithColorCache(true))
	term, ct := env["TERM"], env["COLORTERM"]
	switch {
	case term == "" || term == "dumb":
		r.SetColorProfile(termenv.Ascii)
	case ct == "truecolor" || ct == "24bit" || strings.Contains(term, "truecolor") || strings.Contains(term, "direct"):
		r.SetColorProfile(termenv.TrueColor)
	default:
		r.SetColorProfile(termenv.ANSI256)
	}
	return r
}
