package dns

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/thomasilliet/gfw/internal/store"
)

// fakeRepository records DNS worker persistence calls.
type fakeRepository struct {
	hosts      []store.TrustedHostname
	listErr    error
	savedID    string
	saved      []string
	savedErr   error
	saveReturn error
}

func (f *fakeRepository) ListActiveHostnames(context.Context) ([]store.TrustedHostname, error) {
	return f.hosts, f.listErr
}
func (f *fakeRepository) SaveResolution(_ context.Context, id string, ips []string, err error) error {
	f.savedID, f.saved, f.savedErr = id, ips, err
	return f.saveReturn
}

// TestResolve covers filtering, sorting, failures, and empty answers.
func TestResolve(t *testing.T) {
	repo := &fakeRepository{}
	w := New(repo, time.Hour, time.Second)
	w.lookup = func(context.Context, string, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.2"), net.ParseIP("2001:db8::1"), net.ParseIP("10.0.0.1")}, nil
	}
	if err := w.Resolve(context.Background(), store.TrustedHostname{ID: "id", Hostname: "example.test"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.saved, []string{"10.0.0.1", "10.0.0.2"}) {
		t.Fatalf("unexpected IPs: %v", repo.saved)
	}
	w.lookup = func(context.Context, string, string) ([]net.IP, error) { return nil, errors.New("lookup") }
	_ = w.Resolve(context.Background(), store.TrustedHostname{ID: "id"})
	if repo.savedErr == nil {
		t.Fatal("expected saved lookup error")
	}
	w.lookup = func(context.Context, string, string) ([]net.IP, error) { return nil, nil }
	_ = w.Resolve(context.Background(), store.TrustedHostname{ID: "id", Hostname: "empty"})
	if repo.savedErr == nil {
		t.Fatal("expected empty-answer error")
	}
}

// TestResolveAllAndLifecycle covers repository errors, iteration, start, and close.
func TestResolveAllAndLifecycle(t *testing.T) {
	repo := &fakeRepository{listErr: errors.New("list")}
	w := New(repo, time.Millisecond, time.Second)
	if err := w.ResolveAll(context.Background()); err == nil {
		t.Fatal("expected list error")
	}
	repo.listErr, repo.hosts = nil, []store.TrustedHostname{{ID: "one", Hostname: "one"}}
	w.lookup = func(context.Context, string, string) ([]net.IP, error) { return []net.IP{net.ParseIP("10.0.0.1")}, nil }
	if err := w.ResolveAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	w.Close()
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)
	time.Sleep(5 * time.Millisecond)
	cancel()
	w.Close()
}
