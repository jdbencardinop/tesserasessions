package statusprovider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/core"
)

var fixedTime = time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)

type fakeScanner struct {
	name string
	scan func(context.Context) core.ScanResult
}

func (f fakeScanner) Name() string {
	return f.name
}

func (f fakeScanner) Scan(ctx context.Context) core.ScanResult {
	return f.scan(ctx)
}

func completeScanner(name string, runtimes ...core.RuntimeInstance) adapters.Scanner {
	return fakeScanner{
		name: name,
		scan: func(context.Context) core.ScanResult {
			return core.ScanResult{
				Source:           name,
				Runtimes:         runtimes,
				SnapshotComplete: true,
			}
		},
	}
}

func testService(scanners ...adapters.Scanner) *Service {
	service := NewService(scanners)
	service.Now = func() time.Time { return fixedTime }
	service.ResolveGit = func(context.Context, string) (string, string) {
		return "", ""
	}
	return service
}

func TestServiceUsesBranchHintForSharedCheckoutPath(t *testing.T) {
	service := testService(completeScanner("herdr", core.RuntimeInstance{
		ID:          "herdr-1",
		Backend:     "herdr",
		NativeID:    "w1:p1",
		ProjectPath: "/repo",
		Status:      core.StatusNeedsAttention,
	}))
	service.ResolveGit = func(context.Context, string) (string, string) {
		return "/repo", "feature/a"
	}

	metadata := json.RawMessage(`{"workspace_id":"ws-1"}`)
	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "a", Path: "/repo", RepoRoot: "/repo", GitBranch: "feature/a", Metadata: metadata},
			{QueryID: "b", Path: "/repo", RepoRoot: "/repo", GitBranch: "feature/b"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	first := response.Results[0]
	if first.RuntimePresence != RuntimePresent || first.AgentState != AgentBlocked {
		t.Fatalf("unexpected first aggregate: %+v", first)
	}
	if first.Match == nil || first.Match.Type != "repo_branch" {
		t.Fatalf("expected repo_branch match, got %+v", first.Match)
	}
	if len(first.Runtimes) != 1 {
		t.Fatalf("expected one raw runtime, got %d", len(first.Runtimes))
	}
	if string(first.Metadata) != string(metadata) {
		t.Fatalf("metadata was not echoed: %s", first.Metadata)
	}

	second := response.Results[1]
	if second.RuntimePresence != RuntimeAbsent || len(second.Runtimes) != 0 {
		t.Fatalf("branch b should be absent, got %+v", second)
	}
}

func TestBranchHintDoesNotOverrideSiblingPathContainment(t *testing.T) {
	service := testService(completeScanner(
		"herdr",
		core.RuntimeInstance{
			ID:          "frontend-runtime",
			Backend:     "herdr",
			NativeID:    "frontend",
			ProjectPath: "/repo/frontend",
			Status:      core.StatusWorking,
		},
		core.RuntimeInstance{
			ID:          "backend-runtime",
			Backend:     "herdr",
			NativeID:    "backend",
			ProjectPath: "/repo/backend",
			Status:      core.StatusIdle,
		},
	))
	service.ResolveGit = func(context.Context, string) (string, string) {
		return "/repo", "main"
	}

	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "frontend", Path: "/repo/frontend", RepoRoot: "/repo", GitBranch: "main"},
			{QueryID: "backend", Path: "/repo/backend", RepoRoot: "/repo", GitBranch: "main"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Results[0].Runtimes; len(got) != 1 || got[0].NativeID != "frontend" {
		t.Fatalf("frontend received wrong runtimes: %+v", got)
	}
	if got := response.Results[1].Runtimes; len(got) != 1 || got[0].NativeID != "backend" {
		t.Fatalf("backend received wrong runtimes: %+v", got)
	}
}

func TestServiceAssignsRuntimeToDeepestPath(t *testing.T) {
	service := testService(completeScanner("tmux", core.RuntimeInstance{
		ID:          "tmux-1",
		Backend:     "tmux",
		NativeID:    "%1",
		ProjectPath: "/repo/worktrees/auth/src",
		Status:      core.StatusIdle,
	}))

	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "repo", Path: "/repo"},
			{QueryID: "auth", Path: "/repo/worktrees/auth"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Results[0].RuntimePresence != RuntimeAbsent {
		t.Fatalf("parent query should not receive the runtime: %+v", response.Results[0])
	}
	child := response.Results[1]
	if child.RuntimePresence != RuntimePresent || child.AgentState != AgentReady {
		t.Fatalf("unexpected child aggregate: %+v", child)
	}
	if child.Match == nil || child.Match.Type != "descendant_path" {
		t.Fatalf("unexpected match: %+v", child.Match)
	}
}

func TestServiceAggregatesRawRuntimeStates(t *testing.T) {
	service := testService(completeScanner(
		"herdr",
		core.RuntimeInstance{ID: "one", Backend: "herdr", NativeID: "one", ProjectPath: "/repo", Status: core.StatusIdle},
		core.RuntimeInstance{ID: "two", Backend: "herdr", NativeID: "two", ProjectPath: "/repo", Status: core.StatusNeedsAttention},
	))

	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries:       []Query{{QueryID: "repo", Path: "/repo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	result := response.Results[0]
	if result.RuntimePresence != RuntimePresent || result.AgentState != AgentBlocked {
		t.Fatalf("unexpected aggregate: %+v", result)
	}
	if len(result.Runtimes) != 2 {
		t.Fatalf("expected both raw runtimes, got %d", len(result.Runtimes))
	}
}

func TestServiceAggregatesStaleRuntimePresence(t *testing.T) {
	tests := []struct {
		name     string
		runtimes []core.RuntimeInstance
		want     RuntimePresence
	}{
		{
			name: "stale only",
			runtimes: []core.RuntimeInstance{
				{ID: "stale", Backend: "herdr", NativeID: "stale", ProjectPath: "/repo", Status: core.StatusStale},
			},
			want: RuntimeStale,
		},
		{
			name: "present beats stale",
			runtimes: []core.RuntimeInstance{
				{ID: "stale", Backend: "herdr", NativeID: "stale", ProjectPath: "/repo", Status: core.StatusStale},
				{ID: "live", Backend: "herdr", NativeID: "live", ProjectPath: "/repo", Status: core.StatusIdle},
			},
			want: RuntimePresent,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := testService(completeScanner("herdr", test.runtimes...))
			response, err := service.Query(context.Background(), Request{
				SchemaVersion: SchemaVersion,
				Queries:       []Query{{QueryID: "repo", Path: "/repo"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := response.Results[0].RuntimePresence; got != test.want {
				t.Fatalf("presence = %q, want %q", got, test.want)
			}
		})
	}
}

func TestServiceReturnsMixedQueryErrors(t *testing.T) {
	service := testService(completeScanner("tmux", core.RuntimeInstance{
		ID:          "tmux-1",
		Backend:     "tmux",
		NativeID:    "%1",
		ProjectPath: "/repo",
		Status:      core.StatusUnknown,
	}))

	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "valid", Path: "/repo"},
			{QueryID: "invalid", Path: "relative/path"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Results[0].RuntimePresence != RuntimePresent {
		t.Fatalf("valid query did not succeed: %+v", response.Results[0])
	}
	invalid := response.Results[1]
	if invalid.RuntimePresence != RuntimeUnknown || len(invalid.Errors) != 1 || invalid.Errors[0].Code != "invalid_path" {
		t.Fatalf("unexpected invalid result: %+v", invalid)
	}
}

func TestServiceReportsIncompleteCoverage(t *testing.T) {
	service := testService(fakeScanner{
		name: "herdr",
		scan: func(context.Context) core.ScanResult {
			return core.ScanResult{Source: "herdr", Skipped: true, Message: "herdr not found"}
		},
	})

	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries:       []Query{{QueryID: "repo", Path: "/repo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	result := response.Results[0]
	if result.RuntimePresence != RuntimeUnknown {
		t.Fatalf("incomplete coverage must be unknown: %+v", result)
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "coverage_incomplete" {
		t.Fatalf("missing coverage error: %+v", result.Errors)
	}
	if response.Providers[0].Available {
		t.Fatal("provider should be unavailable")
	}
}

func TestServiceBoundsProviderTimeout(t *testing.T) {
	service := testService(fakeScanner{
		name: "slow",
		scan: func(ctx context.Context) core.ScanResult {
			<-ctx.Done()
			return core.ScanResult{Source: "slow", Err: ctx.Err()}
		},
	})
	service.Timeout = 10 * time.Millisecond

	start := time.Now()
	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries:       []Query{{QueryID: "repo", Path: "/repo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("probe exceeded bound: %s", elapsed)
	}
	if response.Providers[0].Available || response.Providers[0].Error == "" {
		t.Fatalf("timeout not surfaced: %+v", response.Providers[0])
	}
}

func TestServiceResolvesSymlinkQueryPath(t *testing.T) {
	root := t.TempDir()
	realPath := filepath.Join(root, "real")
	linkPath := filepath.Join(root, "link")
	if err := os.MkdirAll(filepath.Join(realPath, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	service := testService(completeScanner("tmux", core.RuntimeInstance{
		ID:          "tmux-1",
		Backend:     "tmux",
		NativeID:    "%1",
		ProjectPath: filepath.Join(realPath, "src"),
		Status:      core.StatusIdle,
	}))
	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries:       []Query{{QueryID: "linked", Path: linkPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Results[0].RuntimePresence != RuntimePresent {
		t.Fatalf("symlink query did not match: %+v", response.Results[0])
	}
}

func TestServiceRejectsInvalidEnvelope(t *testing.T) {
	service := testService()
	_, err := service.Query(context.Background(), Request{
		SchemaVersion: 99,
		Queries:       []Query{{QueryID: "repo", Path: "/repo"}},
	})
	if err == nil {
		t.Fatal("expected schema error")
	}
	_, err = service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "duplicate", Path: "/one"},
			{QueryID: "duplicate", Path: "/two"},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate query id error")
	}
}

func TestAmbiguousEqualSpecificityRemainsUnknown(t *testing.T) {
	service := testService(completeScanner("tmux", core.RuntimeInstance{
		ID:          "tmux-1",
		Backend:     "tmux",
		NativeID:    "%1",
		ProjectPath: "/repo",
		Status:      core.StatusIdle,
	}))
	response, err := service.Query(context.Background(), Request{
		SchemaVersion: SchemaVersion,
		Queries: []Query{
			{QueryID: "one", Path: "/repo"},
			{QueryID: "two", Path: "/repo"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range response.Results {
		if result.RuntimePresence != RuntimeUnknown || len(result.Errors) == 0 || result.Errors[0].Code != "ambiguous_match" {
			t.Fatalf("unexpected ambiguous result: %+v", result)
		}
	}
}
