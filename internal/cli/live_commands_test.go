package cli

import (
	"strings"
	"testing"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestSendCommandUsesHerdrPrompt(t *testing.T) {
	got, err := sendCommand(core.RuntimeInstance{Backend: "herdr", NativeID: "agent-1"}, "continue")
	if err != nil {
		t.Fatal(err)
	}
	want := "herdr agent prompt 'agent-1' 'continue'"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReadCommandTmux(t *testing.T) {
	got, err := readCommand(core.RuntimeInstance{Backend: "tmux", NativeID: "%1"}, 40)
	if err != nil {
		t.Fatal(err)
	}
	want := "tmux capture-pane -p -t '%1' -S -40"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRunCommandHerdrSplit(t *testing.T) {
	got, err := runCommand(
		core.RuntimeInstance{Backend: "herdr", NativeID: "agent-1", Surface: "w1:p1", ProjectPath: "/repo"},
		core.Session{},
		"go test ./...",
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"herdr pane split 'w1:p1'",
		"--cwd '/repo'",
		"herdr pane run \"$pane\" 'go test ./...'",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("command %q missing %q", got, want)
		}
	}
}
