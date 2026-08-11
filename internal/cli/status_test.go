package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/core"
	"github.com/jdbencardinop/tesserasessions/internal/statusprovider"
)

type statusScannerStub struct{}

func (statusScannerStub) Name() string {
	return "stub"
}

func (statusScannerStub) Scan(context.Context) core.ScanResult {
	return core.ScanResult{
		Source:           "stub",
		SnapshotComplete: true,
		Runtimes: []core.RuntimeInstance{{
			ID:          "stub-1",
			Backend:     "stub",
			NativeID:    "native-1",
			ProjectPath: "/repo",
			Status:      core.StatusWorking,
		}},
	}
}

func TestStatusCommandUsesVersionedLowercaseJSONWithoutDatabase(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("TSS_DATA_DIR", dataDir)

	service := statusprovider.NewService([]adapters.Scanner{statusScannerStub{}})
	service.Now = func() time.Time {
		return time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)
	}
	service.ResolveGit = func(context.Context, string) (string, string) {
		return "", ""
	}

	cmd := newStatusCmd(service)
	cmd.SetArgs([]string{"--json"})
	cmd.SetIn(strings.NewReader(`{"schema_version":1,"queries":[{"query_id":"repo","path":"/repo"}]}`))
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status failed: %v\n%s", err, stderr.String())
	}

	if strings.Contains(stdout.String(), "SchemaVersion") || !strings.Contains(stdout.String(), `"schema_version":1`) {
		t.Fatalf("wire JSON is not explicit lowercase schema: %s", stdout.String())
	}
	var response statusprovider.Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Results[0].RuntimePresence != statusprovider.RuntimePresent ||
		response.Results[0].AgentState != statusprovider.AgentWorking {
		t.Fatalf("unexpected response: %+v", response.Results[0])
	}

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("status command created inventory state: %+v", entries)
	}
}

func TestStatusCommandRequiresJSON(t *testing.T) {
	cmd := newStatusCmd(statusprovider.NewService(nil))
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "requires --json") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeStatusRequestRejectsUnknownAndTrailingData(t *testing.T) {
	_, err := decodeStatusRequest(strings.NewReader(`{"schema_version":1,"queries":[],"extra":true}`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	_, err = decodeStatusRequest(strings.NewReader(`{"schema_version":1,"queries":[]} {}`))
	if err == nil {
		t.Fatal("expected trailing value error")
	}
}
