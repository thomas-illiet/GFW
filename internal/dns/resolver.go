package dns

import (
	"context"
	"net"
	"net/netip"
	"sort"
	"time"

	"github.com/thomasilliet/gfw/internal/store"
)

type Worker struct {
	store             Repository
	interval, timeout time.Duration
	cancel            context.CancelFunc
	lookup            LookupFunc
}

// Repository defines persistence needed by the DNS worker.
type Repository interface {
	ListActiveHostnames(context.Context) ([]store.TrustedHostname, error)
	SaveResolution(context.Context, string, []string, error) error
}

// LookupFunc resolves addresses for one network and hostname.
type LookupFunc func(context.Context, string, string) ([]net.IP, error)

// New constructs a DNS worker.
func New(s Repository, interval, timeout time.Duration) *Worker {
	return &Worker{store: s, interval: interval, timeout: timeout, lookup: net.DefaultResolver.LookupIP}
}

// Start launches immediate and periodic trusted-hostname resolution.
func (w *Worker) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		w.ResolveAll(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.ResolveAll(ctx)
			}
		}
	}()
}

// Close stops the worker when it has been started.
func (w *Worker) Close() {
	if w.cancel != nil {
		w.cancel()
	}
}

// ResolveAll resolves every active trusted hostname in order.
func (w *Worker) ResolveAll(ctx context.Context) error {
	hosts, err := w.store.ListActiveHostnames(ctx)
	if err != nil {
		return err
	}
	for _, h := range hosts {
		if err := w.Resolve(ctx, h); err != nil {
			return err
		}
	}
	return nil
}

// Resolve refreshes one trusted hostname and persists success or failure.
func (w *Worker) Resolve(ctx context.Context, h store.TrustedHostname) error {
	resolveCtx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	addrs, err := w.lookup(resolveCtx, "ip4", h.Hostname)
	ips := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		a, ok := netip.AddrFromSlice(addr)
		a = a.Unmap()
		if ok && a.Is4() {
			ips = append(ips, a.String())
		}
	}
	sort.Strings(ips)
	if err == nil && len(ips) == 0 {
		err = &net.DNSError{Err: "no IPv4 records", Name: h.Hostname}
	}
	return w.store.SaveResolution(ctx, h.ID, ips, err)
}
