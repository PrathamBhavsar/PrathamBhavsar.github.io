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
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
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
				return newModel(site, pty.Window.Width, pty.Window.Height), []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
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
