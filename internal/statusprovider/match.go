package statusprovider

import (
	"fmt"
	"path/filepath"
	stdruntime "runtime"
	"strings"
)

const (
	matchRankDescendant = 1_000_000
	matchRankExact      = 2_000_000
	matchRankRepoBranch = 3_000_000
)

type preparedQuery struct {
	index       int
	query       Query
	path        string
	pathKey     string
	repoRoot    string
	repoRootKey string
	errors      []QueryError
}

type preparedObservation struct {
	RuntimeObservation
	pathKey     string
	repoRootKey string
}

type matchCandidate struct {
	queryIndex  int
	matchType   string
	matchedPath string
	rank        int
}

func matchQueries(queries []Query, providers []ProviderStatus, observations []RuntimeObservation) []Result {
	prepared := prepareQueries(queries)
	results := make([]Result, len(queries))
	for index, query := range prepared {
		results[index] = Result{
			QueryID:         query.query.QueryID,
			Metadata:        query.query.Metadata,
			RuntimePresence: RuntimeUnknown,
			AgentState:      AgentUnknown,
			Runtimes:        []RuntimeObservation{},
			Errors:          append([]QueryError{}, query.errors...),
		}
	}

	preparedObservations := prepareObservations(observations)
	bestMatchRank := make([]int, len(results))
	for _, observation := range preparedObservations {
		candidates := candidatesForObservation(prepared, observation)
		if len(candidates) == 0 {
			continue
		}
		best := candidates[0].rank
		var winners []matchCandidate
		for _, candidate := range candidates {
			if candidate.rank > best {
				best = candidate.rank
				winners = winners[:0]
			}
			if candidate.rank == best {
				winners = append(winners, candidate)
			}
		}
		if len(winners) != 1 {
			for _, winner := range winners {
				appendQueryError(&results[winner.queryIndex], QueryError{
					Code:    "ambiguous_match",
					Message: fmt.Sprintf("runtime %s matches multiple queries with equal specificity", observation.ObservationID),
				})
			}
			continue
		}

		winner := winners[0]
		result := &results[winner.queryIndex]
		result.Runtimes = append(result.Runtimes, observation.RuntimeObservation)
		if winner.rank > bestMatchRank[winner.queryIndex] {
			result.Match = &MatchEvidence{Type: winner.matchType, MatchedPath: winner.matchedPath}
			bestMatchRank[winner.queryIndex] = winner.rank
		}
	}

	allProvidersAvailable := len(providers) > 0
	var unavailable []string
	for _, provider := range providers {
		if !provider.Available {
			allProvidersAvailable = false
			unavailable = append(unavailable, provider.Name)
		}
	}

	for index := range results {
		result := &results[index]
		if len(result.Runtimes) > 0 {
			result.RuntimePresence = aggregatePresence(result.Runtimes)
			result.AgentState = aggregateAgentState(result.Runtimes)
			continue
		}
		if len(result.Errors) > 0 {
			continue
		}
		if allProvidersAvailable {
			result.RuntimePresence = RuntimeAbsent
			continue
		}
		if len(unavailable) > 0 {
			appendQueryError(result, QueryError{
				Code:    "coverage_incomplete",
				Message: "unavailable providers: " + strings.Join(unavailable, ", "),
			})
		}
	}
	return results
}

func prepareQueries(queries []Query) []preparedQuery {
	prepared := make([]preparedQuery, len(queries))
	for index, query := range queries {
		item := preparedQuery{index: index, query: query}
		var pathErr, repoErr error
		item.path, item.pathKey, pathErr = canonicalPathParts(query.Path)
		item.repoRoot, item.repoRootKey, repoErr = canonicalPathParts(query.RepoRoot)

		if query.Path != "" && pathErr != nil {
			item.errors = append(item.errors, QueryError{Code: "invalid_path", Message: pathErr.Error()})
		}
		if query.RepoRoot != "" && repoErr != nil {
			item.errors = append(item.errors, QueryError{Code: "invalid_repo_root", Message: repoErr.Error()})
		}
		if query.Path == "" && (query.RepoRoot == "" || strings.TrimSpace(query.GitBranch) == "") {
			item.errors = append(item.errors, QueryError{
				Code:    "insufficient_match_evidence",
				Message: "path or repo_root plus git_branch is required",
			})
		}
		prepared[index] = item
	}
	return prepared
}

func prepareObservations(observations []RuntimeObservation) []preparedObservation {
	prepared := make([]preparedObservation, 0, len(observations))
	for _, observation := range observations {
		_, pathKey, _ := canonicalPathParts(observation.Path)
		_, repoKey, _ := canonicalPathParts(observation.RepoRoot)
		prepared = append(prepared, preparedObservation{
			RuntimeObservation: observation,
			pathKey:            pathKey,
			repoRootKey:        repoKey,
		})
	}
	return prepared
}

func candidatesForObservation(queries []preparedQuery, observation preparedObservation) []matchCandidate {
	var candidates []matchCandidate
	for _, query := range queries {
		if len(query.errors) > 0 {
			continue
		}
		pathCompatible := query.pathKey == "" ||
			observation.pathKey == query.pathKey ||
			pathWithin(observation.pathKey, query.pathKey)
		if query.repoRootKey != "" &&
			strings.TrimSpace(query.query.GitBranch) != "" &&
			observation.repoRootKey == query.repoRootKey &&
			observation.GitBranch == strings.TrimSpace(query.query.GitBranch) &&
			pathCompatible {
			candidates = append(candidates, matchCandidate{
				queryIndex:  query.index,
				matchType:   "repo_branch",
				matchedPath: observation.Path,
				rank:        matchRankRepoBranch + pathSpecificity(query.pathKey),
			})
			continue
		}
		if query.pathKey == "" || observation.pathKey == "" {
			continue
		}
		if observation.pathKey == query.pathKey {
			candidates = append(candidates, matchCandidate{
				queryIndex:  query.index,
				matchType:   "exact_path",
				matchedPath: observation.Path,
				rank:        matchRankExact + pathSpecificity(query.pathKey),
			})
			continue
		}
		if pathWithin(observation.pathKey, query.pathKey) {
			candidates = append(candidates, matchCandidate{
				queryIndex:  query.index,
				matchType:   "descendant_path",
				matchedPath: observation.Path,
				rank:        matchRankDescendant + pathSpecificity(query.pathKey),
			})
		}
	}
	return candidates
}

func canonicalPath(path string) (string, error) {
	display, _, err := canonicalPathParts(path)
	return display, err
}

func canonicalPathParts(path string) (string, string, error) {
	if strings.TrimSpace(path) == "" {
		return "", "", nil
	}
	if !filepath.IsAbs(path) {
		return "", "", fmt.Errorf("path must be absolute: %s", path)
	}
	display := filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(display); err == nil {
		display = filepath.Clean(resolved)
	}
	key := display
	if stdruntime.GOOS == "darwin" || stdruntime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return display, key, nil
}

func pathWithin(child, parent string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil || relative == "." || filepath.IsAbs(relative) {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func pathSpecificity(path string) int {
	return len(path)
}

func aggregatePresence(observations []RuntimeObservation) RuntimePresence {
	best := RuntimeUnknown
	for _, observation := range observations {
		switch observation.RuntimePresence {
		case RuntimePresent:
			return RuntimePresent
		case RuntimeStale:
			best = RuntimeStale
		}
	}
	return best
}

func aggregateAgentState(observations []RuntimeObservation) AgentState {
	best := AgentUnknown
	bestRank := 0
	for _, observation := range observations {
		rank := agentStateRank(observation.AgentState)
		if rank > bestRank {
			best = observation.AgentState
			bestRank = rank
		}
	}
	return best
}

func agentStateRank(state AgentState) int {
	switch state {
	case AgentBlocked:
		return 5
	case AgentWorking:
		return 4
	case AgentReady:
		return 3
	case AgentDone:
		return 2
	default:
		return 1
	}
}

func appendQueryError(result *Result, queryError QueryError) {
	for _, existing := range result.Errors {
		if existing.Code == queryError.Code && existing.Message == queryError.Message {
			return
		}
	}
	result.Errors = append(result.Errors, queryError)
}
