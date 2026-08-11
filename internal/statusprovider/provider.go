package statusprovider

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/core"
)

const (
	defaultProbeTimeout = 2 * time.Second
	defaultFreshFor     = 10 * time.Second
)

type GitResolver func(context.Context, string) (repoRoot, branch string)

type Service struct {
	Scanners   []adapters.Scanner
	Timeout    time.Duration
	FreshFor   time.Duration
	Now        func() time.Time
	ResolveGit GitResolver
}

func NewService(scanners []adapters.Scanner) *Service {
	return &Service{
		Scanners:   scanners,
		Timeout:    defaultProbeTimeout,
		FreshFor:   defaultFreshFor,
		Now:        time.Now,
		ResolveGit: resolveGit,
	}
}

func (s *Service) Query(ctx context.Context, request Request) (Response, error) {
	if err := validateEnvelope(request); err != nil {
		return Response{}, err
	}
	s.applyDefaults()

	probes := s.probe(ctx)
	observedAt := s.Now().UTC()
	providers := make([]ProviderStatus, 0, len(probes))
	var observations []RuntimeObservation
	for _, probe := range probes {
		providers = append(providers, probe.provider)
		observations = append(observations, probe.observations...)
	}

	return Response{
		SchemaVersion: SchemaVersion,
		ObservedAt:    observedAt,
		FreshUntil:    observedAt.Add(s.FreshFor),
		Providers:     providers,
		Results:       matchQueries(request.Queries, providers, observations),
	}, nil
}

func (s *Service) applyDefaults() {
	if s.Timeout <= 0 {
		s.Timeout = defaultProbeTimeout
	}
	if s.FreshFor <= 0 {
		s.FreshFor = defaultFreshFor
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.ResolveGit == nil {
		s.ResolveGit = resolveGit
	}
}

type probeOutput struct {
	index        int
	provider     ProviderStatus
	observations []RuntimeObservation
}

func (s *Service) probe(ctx context.Context) []probeOutput {
	outputs := make(chan probeOutput, len(s.Scanners))
	var wg sync.WaitGroup

	for index, scanner := range s.Scanners {
		wg.Add(1)
		go func(index int, scanner adapters.Scanner) {
			defer wg.Done()
			outputs <- s.probeOne(ctx, index, scanner)
		}(index, scanner)
	}

	wg.Wait()
	close(outputs)

	result := make([]probeOutput, 0, len(s.Scanners))
	for output := range outputs {
		result = append(result, output)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].index < result[j].index
	})
	return result
}

func (s *Service) probeOne(parent context.Context, index int, scanner adapters.Scanner) probeOutput {
	ctx, cancel := context.WithTimeout(parent, s.Timeout)
	defer cancel()

	scan := scanner.Scan(ctx)
	observedAt := s.Now().UTC()
	provider := ProviderStatus{
		Name:       scanner.Name(),
		Available:  scan.Err == nil && (!scan.Skipped || scan.SnapshotComplete),
		ObservedAt: observedAt,
		FreshUntil: observedAt.Add(s.FreshFor),
	}
	if scan.Err != nil {
		provider.Error = scan.Err.Error()
	} else if scan.Skipped && !scan.SnapshotComplete {
		provider.Error = scan.Message
	}

	observations := make([]RuntimeObservation, 0, len(scan.Runtimes))
	for _, runtime := range scan.Runtimes {
		observations = append(observations, s.runtimeObservation(ctx, runtime, observedAt))
	}
	return probeOutput{index: index, provider: provider, observations: observations}
}

func (s *Service) runtimeObservation(ctx context.Context, runtime core.RuntimeInstance, observedAt time.Time) RuntimeObservation {
	path, _ := canonicalPath(runtime.ProjectPath)
	repoRoot, branch := "", ""
	if path != "" {
		repoRoot, branch = s.ResolveGit(ctx, path)
		repoRoot, _ = canonicalPath(repoRoot)
	}

	presence := RuntimePresent
	if runtime.Status == core.StatusStale {
		presence = RuntimeStale
	}

	observationID := runtime.ID
	if observationID == "" {
		observationID = runtime.Backend + ":" + runtime.NativeID
	}
	return RuntimeObservation{
		ObservationID:   observationID,
		Backend:         runtime.Backend,
		NativeID:        runtime.NativeID,
		Path:            path,
		RepoRoot:        repoRoot,
		GitBranch:       branch,
		RuntimePresence: presence,
		AgentState:      agentState(runtime.Status),
		ObservedAt:      observedAt,
		ExpiresAt:       observedAt.Add(s.FreshFor),
	}
}

func agentState(status string) AgentState {
	switch status {
	case core.StatusNeedsAttention:
		return AgentBlocked
	case core.StatusWorking:
		return AgentWorking
	case core.StatusIdle:
		return AgentReady
	case core.StatusDone:
		return AgentDone
	default:
		return AgentUnknown
	}
}

func resolveGit(ctx context.Context, path string) (string, string) {
	root := gitOutput(ctx, path, "rev-parse", "--show-toplevel")
	if root == "" {
		return "", ""
	}
	branch := gitOutput(ctx, path, "symbolic-ref", "--quiet", "--short", "HEAD")
	return root, branch
}

func gitOutput(ctx context.Context, path string, args ...string) string {
	commandArgs := append([]string{"-C", path}, args...)
	out, err := exec.CommandContext(ctx, "git", commandArgs...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func validateEnvelope(request Request) error {
	if request.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %d (expected %d)", request.SchemaVersion, SchemaVersion)
	}
	seen := make(map[string]struct{}, len(request.Queries))
	for index, query := range request.Queries {
		id := strings.TrimSpace(query.QueryID)
		if id == "" {
			return fmt.Errorf("queries[%d].query_id is required", index)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate query_id %q", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}
