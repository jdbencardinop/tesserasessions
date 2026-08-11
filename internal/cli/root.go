package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/buildinfo"
	"github.com/jdbencardinop/tesserasessions/internal/config"
	"github.com/jdbencardinop/tesserasessions/internal/core"
	"github.com/jdbencardinop/tesserasessions/internal/store"
	"github.com/jdbencardinop/tesserasessions/internal/summarize"
	"github.com/spf13/cobra"
)

type appContext struct {
	configPath string
	dbPath     string
}

func mergeSessions(primary, secondary []core.Session) []core.Session {
	seen := make(map[string]bool)
	var out []core.Session
	for _, sess := range primary {
		if seen[sess.ID] {
			continue
		}
		seen[sess.ID] = true
		out = append(out, sess)
	}
	for _, sess := range secondary {
		if seen[sess.ID] {
			continue
		}
		seen[sess.ID] = true
		out = append(out, sess)
	}
	return out
}

func Execute() error {
	app := &appContext{}
	root := &cobra.Command{
		Use:     "tss",
		Short:   "Inventory local coding-agent sessions",
		Long:    "tss inventories local coding-agent sessions and connects them to Herdr or tmux when available.",
		Version: buildinfo.String(),
	}
	root.SetVersionTemplate("tss {{.Version}}\n")
	root.PersistentFlags().StringVar(&app.configPath, "config", "", "config file path")
	root.PersistentFlags().StringVar(&app.dbPath, "db", "", "SQLite database path")
	root.AddCommand(
		app.scanCmd(),
		app.statusCmd(),
		app.listCmd(),
		app.showCmd(),
		app.doctorCmd(),
		app.summarizeCmd(),
		app.searchCmd(),
		app.titleCmd(),
		app.markCmd(),
		app.pinCmd(),
		app.tagCmd(),
		app.attachCmd(),
		app.openCmd(),
		app.sendCmd(),
		app.readCmd(),
		app.runCmd(),
	)
	return root.Execute()
}

func (a *appContext) load() (config.Config, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return cfg, err
	}
	if a.dbPath != "" {
		cfg.Database = a.dbPath
	}
	return cfg, nil
}

func (a *appContext) openStore(cfg config.Config) (*store.Store, error) {
	if err := config.EnsureDirs(cfg); err != nil {
		return nil, err
	}
	return store.Open(cfg.Database)
}

func (a *appContext) scanCmd() *cobra.Command {
	var source string
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Refresh the local session inventory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			ctx := cmd.Context()
			type report struct {
				Source   string `json:"source"`
				Sessions int    `json:"sessions"`
				Runtimes int    `json:"runtimes"`
				Skipped  bool   `json:"skipped"`
				Message  string `json:"message,omitempty"`
				Error    string `json:"error,omitempty"`
			}
			var reports []report
			for _, scanner := range adapters.DefaultScanners(cfg) {
				if source != "" && scanner.Name() != source {
					continue
				}
				res := scanner.Scan(ctx)
				r := report{Source: res.Source, Sessions: len(res.Sessions), Runtimes: len(res.Runtimes), Skipped: res.Skipped, Message: res.Message}
				if res.Err != nil {
					r.Error = res.Err.Error()
					_ = db.UpsertSource(ctx, res.Source, "scanner", sourcePath(cfg, res.Source), r.Error)
					reports = append(reports, r)
					continue
				}
				for _, sess := range res.Sessions {
					if err := db.UpsertSession(ctx, sess); err != nil {
						return err
					}
				}
				if err := persistScanRuntimes(ctx, db, res); err != nil {
					return err
				}
				_ = db.UpsertSource(ctx, res.Source, "scanner", sourcePath(cfg, res.Source), "")
				reports = append(reports, r)
			}
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(reports)
			}
			for _, r := range reports {
				if r.Error != "" {
					fmt.Printf("%s: error: %s\n", r.Source, r.Error)
					continue
				}
				if r.Skipped {
					fmt.Printf("%s: skipped (%s)\n", r.Source, r.Message)
					continue
				}
				fmt.Printf("%s: %d session(s), %d runtime(s)\n", r.Source, r.Sessions, r.Runtimes)
			}
			count, err := db.CountSessions(ctx)
			if err == nil {
				fmt.Printf("Inventory: %d session(s)\n", count)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "scan one source (claude, copilot, herdr, tmux)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func persistScanRuntimes(ctx context.Context, db *store.Store, result core.ScanResult) error {
	if result.Err == nil && result.SnapshotComplete {
		return db.ReplaceRuntimes(ctx, result.Source, result.Runtimes)
	}
	for _, runtime := range result.Runtimes {
		if err := db.UpsertRuntime(ctx, runtime); err != nil {
			return err
		}
	}
	return nil
}

func (a *appContext) listCmd() *cobra.Command {
	var filter store.Filter
	var jsonOutput bool
	var groupProject bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List indexed sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			if groupProject && filter.Sort == "" {
				filter.Sort = "project"
			}
			sessions, err := db.ListSessions(cmd.Context(), filter)
			if err != nil {
				return err
			}
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(sessions)
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found. Run: tss scan")
				return nil
			}
			return printSessions(sessions, groupProject)
		},
	}
	cmd.Flags().StringVar(&filter.Source, "source", "", "filter by source")
	cmd.Flags().StringVar(&filter.Status, "status", "", "filter by status")
	cmd.Flags().StringVarP(&filter.Query, "query", "q", "", "filter by title, summary, project, native id, agent, or tags")
	cmd.Flags().StringVar(&filter.Tag, "tag", "", "filter by tag")
	cmd.Flags().StringVar(&filter.Sort, "sort", "", "sort by last, title, project, source, or status")
	cmd.Flags().IntVar(&filter.Limit, "limit", 100, "maximum sessions to print")
	cmd.Flags().BoolVar(&filter.PinnedOnly, "pinned", false, "show pinned sessions only")
	cmd.Flags().BoolVar(&groupProject, "group-project", false, "group output by project")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func (a *appContext) showCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "show <session>",
		Short: "Show one session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(sess)
			}
			printSession(sess)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func (a *appContext) doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Report configuration and source health",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			fmt.Printf("Config: %s\n", firstNonEmpty(a.configPath, config.ConfigPath()))
			fmt.Printf("Database: %s\n", cfg.Database)
			fmt.Printf("Default live backend: %s\n", cfg.Live.DefaultBackend)
			fmt.Println()
			fmt.Printf("Claude projects: %s [%s]\n", cfg.Sources.ClaudeProjects, existsLabel(cfg.Sources.ClaudeProjects))
			fmt.Printf("Copilot session-state: %s [%s]\n", cfg.Sources.CopilotSessionState, existsLabel(cfg.Sources.CopilotSessionState))
			fmt.Println()
			for _, tool := range []string{"herdr", "tmux", "sqlite3"} {
				path, err := exec.LookPath(tool)
				if err != nil {
					fmt.Printf("%s: missing\n", tool)
				} else {
					fmt.Printf("%s: %s\n", tool, path)
				}
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			count, err := db.CountSessions(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("\nInventory sessions: %d\n", count)
			return nil
		},
	}
}

func (a *appContext) summarizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize <session>",
		Short: "Generate a local/extractive title and goal summary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			res := summarize.Local(cmd.Context(), sess)
			if err := db.UpdateSummary(cmd.Context(), sess.ID, "local-extractive", res.Title, res.Summary, res.Status, res.Confidence); err != nil {
				return err
			}
			fmt.Printf("Title: %s\n", res.Title)
			fmt.Printf("Goal: %s\n", res.Summary)
			if res.Status != "" {
				fmt.Printf("Status: %s\n", res.Status)
			}
			return nil
		},
	}
	return cmd
}

func (a *appContext) searchCmd() *cobra.Command {
	var filter store.Filter
	var jsonOutput bool
	var includeContent bool
	var useFZF bool
	var showSelected bool
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search sessions by metadata, summary, and optionally content",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				filter.Query = args[0]
			}
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sessions, err := db.ListSessions(cmd.Context(), filter)
			if err != nil {
				return err
			}
			if includeContent && filter.Query != "" {
				contentFilter := filter
				contentFilter.Query = ""
				allSessions, err := db.ListSessions(cmd.Context(), contentFilter)
				if err != nil {
					return err
				}
				sessions = mergeSessions(sessions, contentMatches(cmd.Context(), allSessions, filter.Query))
			}
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(sessions)
			}
			if useFZF {
				selected, err := runFZF(sessions)
				if err != nil {
					return err
				}
				if selected == "" {
					return nil
				}
				id := strings.Fields(selected)[0]
				if showSelected {
					sess, err := db.GetSession(cmd.Context(), id)
					if err != nil {
						return err
					}
					printSession(sess)
					return nil
				}
				fmt.Println(id)
				return nil
			}
			return printSessions(sessions, false)
		},
	}
	cmd.Flags().StringVar(&filter.Source, "source", "", "filter by source")
	cmd.Flags().StringVar(&filter.Status, "status", "", "filter by status")
	cmd.Flags().StringVar(&filter.Tag, "tag", "", "filter by tag")
	cmd.Flags().StringVar(&filter.Sort, "sort", "", "sort by last, title, project, source, or status")
	cmd.Flags().IntVar(&filter.Limit, "limit", 100, "maximum sessions to consider")
	cmd.Flags().BoolVar(&filter.PinnedOnly, "pinned", false, "show pinned sessions only")
	cmd.Flags().BoolVar(&includeContent, "content", false, "also search lazy transcript/content snippets")
	cmd.Flags().BoolVar(&useFZF, "fzf", false, "select interactively with fzf")
	cmd.Flags().BoolVar(&showSelected, "show", false, "show the selected session when using --fzf")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func (a *appContext) titleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "title <session> <title>",
		Short: "Set a manual session title",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, sess, err := a.sessionForMutation(cmd, args[0])
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			if err := db.UpdateTitle(cmd.Context(), sess.ID, args[1]); err != nil {
				return err
			}
			fmt.Printf("Updated title for %s\n", sess.ID)
			return nil
		},
	}
}

func (a *appContext) markCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mark <session> <status>",
		Short: "Set a session status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !validStatus(args[1]) {
				return fmt.Errorf("invalid status %q", args[1])
			}
			db, sess, err := a.sessionForMutation(cmd, args[0])
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			if err := db.UpdateStatus(cmd.Context(), sess.ID, args[1]); err != nil {
				return err
			}
			fmt.Printf("Marked %s as %s\n", sess.ID, args[1])
			return nil
		},
	}
}

func (a *appContext) pinCmd() *cobra.Command {
	var unpin bool
	cmd := &cobra.Command{
		Use:   "pin <session>",
		Short: "Pin or unpin a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, sess, err := a.sessionForMutation(cmd, args[0])
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			if err := db.UpdatePinned(cmd.Context(), sess.ID, !unpin); err != nil {
				return err
			}
			if unpin {
				fmt.Printf("Unpinned %s\n", sess.ID)
			} else {
				fmt.Printf("Pinned %s\n", sess.ID)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&unpin, "unpin", false, "remove the pin")
	return cmd
}

func (a *appContext) tagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <session> <tag[,tag...]>",
		Short: "Replace session tags",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, sess, err := a.sessionForMutation(cmd, args[0])
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			if err := db.UpdateTags(cmd.Context(), sess.ID, args[1]); err != nil {
				return err
			}
			fmt.Printf("Updated tags for %s\n", sess.ID)
			return nil
		},
	}
}

func (a *appContext) attachCmd() *cobra.Command {
	var printOnly bool
	var backend string
	cmd := &cobra.Command{
		Use:   "attach <session>",
		Short: "Attach or resume a live session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			command := sess.AttachCommand
			preferred := firstNonEmpty(backend, cfg.Live.DefaultBackend)
			if rt, err := db.RuntimeForSession(cmd.Context(), sess.ID, preferred); err == nil && rt.AttachCommand != "" {
				command = rt.AttachCommand
			}
			if command == "" {
				command = sess.ResumeCommand
			}
			if command == "" {
				return fmt.Errorf("session %s has no attach or resume command", sess.ID)
			}
			if printOnly {
				fmt.Println(command)
				return nil
			}
			return runInteractive(command)
		},
	}
	cmd.Flags().BoolVar(&printOnly, "print", false, "print command instead of running it")
	cmd.Flags().StringVar(&backend, "backend", "", "prefer backend (herdr or tmux)")
	return cmd
}

func (a *appContext) openCmd() *cobra.Command {
	var printOnly bool
	var backend string
	cmd := &cobra.Command{
		Use:   "open <session>",
		Short: "Open a new live workspace or tmux session in the same directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			if sess.ProjectPath == "" {
				return fmt.Errorf("session %s does not have a project path", sess.ID)
			}
			command := openCommand(firstNonEmpty(backend, cfg.Live.DefaultBackend), sess)
			if printOnly {
				fmt.Println(command)
				return nil
			}
			return runInteractive(command)
		},
	}
	cmd.Flags().BoolVar(&printOnly, "print", false, "print command instead of running it")
	cmd.Flags().StringVar(&backend, "backend", "", "backend (herdr or tmux)")
	return cmd
}

func (a *appContext) sendCmd() *cobra.Command {
	var printOnly bool
	var backend string
	cmd := &cobra.Command{
		Use:   "send <session> <text>",
		Short: "Send text to a live agent session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			preferred := firstNonEmpty(backend, cfg.Live.DefaultBackend)
			rt, err := db.RuntimeForSession(cmd.Context(), sess.ID, preferred)
			if err != nil {
				return fmt.Errorf("no live runtime found for %s; run tss scan while the session is open", sess.ID)
			}
			command, err := sendCommand(rt, args[1])
			if err != nil {
				return err
			}
			if printOnly {
				fmt.Println(command)
				return nil
			}
			return runInteractive(command)
		},
	}
	cmd.Flags().BoolVar(&printOnly, "print", false, "print command instead of running it")
	cmd.Flags().StringVar(&backend, "backend", "", "prefer backend (herdr or tmux)")
	return cmd
}

func (a *appContext) readCmd() *cobra.Command {
	var printOnly bool
	var backend string
	var lines int
	cmd := &cobra.Command{
		Use:   "read <session>",
		Short: "Read recent output from a live session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			rt, err := db.RuntimeForSession(cmd.Context(), sess.ID, firstNonEmpty(backend, cfg.Live.DefaultBackend))
			if err != nil {
				return fmt.Errorf("no live runtime found for %s; run tss scan while the session is open", sess.ID)
			}
			command, err := readCommand(rt, lines)
			if err != nil {
				return err
			}
			if printOnly {
				fmt.Println(command)
				return nil
			}
			return runInteractive(command)
		},
	}
	cmd.Flags().IntVar(&lines, "lines", 80, "recent lines to read")
	cmd.Flags().BoolVar(&printOnly, "print", false, "print command instead of running it")
	cmd.Flags().StringVar(&backend, "backend", "", "prefer backend (herdr or tmux)")
	return cmd
}

func (a *appContext) runCmd() *cobra.Command {
	var printOnly bool
	var backend string
	cmd := &cobra.Command{
		Use:   "run <session> -- <command>",
		Short: "Run a command in a new live pane for the session directory",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.load()
			if err != nil {
				return err
			}
			db, err := a.openStore(cfg)
			if err != nil {
				return err
			}
			defer db.Close() //nolint:errcheck
			sess, err := db.GetSession(cmd.Context(), args[0])
			if err != nil {
				return friendlySessionErr(args[0], err)
			}
			rt, err := db.RuntimeForSession(cmd.Context(), sess.ID, firstNonEmpty(backend, cfg.Live.DefaultBackend))
			if err != nil {
				return fmt.Errorf("no live runtime found for %s; run tss scan while the session is open", sess.ID)
			}
			command, err := runCommand(rt, sess, strings.Join(args[1:], " "))
			if err != nil {
				return err
			}
			if printOnly {
				fmt.Println(command)
				return nil
			}
			return runInteractive(command)
		},
	}
	cmd.Flags().BoolVar(&printOnly, "print", false, "print command instead of running it")
	cmd.Flags().StringVar(&backend, "backend", "", "prefer backend (herdr or tmux)")
	return cmd
}

func (a *appContext) sessionForMutation(cmd *cobra.Command, ref string) (*store.Store, core.Session, error) {
	cfg, err := a.load()
	if err != nil {
		return nil, core.Session{}, err
	}
	db, err := a.openStore(cfg)
	if err != nil {
		return nil, core.Session{}, err
	}
	sess, err := db.GetSession(cmd.Context(), ref)
	if err != nil {
		_ = db.Close()
		return nil, core.Session{}, friendlySessionErr(ref, err)
	}
	return db, sess, nil
}

func printSessions(sessions []core.Session, groupProject bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if groupProject {
		current := ""
		for _, sess := range sessions {
			project := core.ProjectName(sess.ProjectPath)
			if project != current {
				if current != "" {
					fmt.Fprintln(w)
				}
				current = project
				fmt.Fprintf(w, "%s\n", project)
				fmt.Fprintln(w, "ID\tPIN\tSOURCE\tSTATUS\tLAST ACTIVITY\tAGENT\tTAGS\tTITLE")
			}
			printSessionRow(w, sess)
		}
		return w.Flush()
	}
	fmt.Fprintln(w, "ID\tPIN\tSOURCE\tSTATUS\tLAST ACTIVITY\tAGENT\tPROJECT\tTAGS\tTITLE")
	for _, sess := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			sess.ID, pinLabel(sess), sess.Source, sess.Status, core.FormatTime(sess.LastActivityAt),
			sess.Agent, core.ProjectName(sess.ProjectPath), sess.Tags, core.Truncate(sess.Title, 52))
	}
	return w.Flush()
}

func printSessionRow(w *tabwriter.Writer, sess core.Session) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		sess.ID, pinLabel(sess), sess.Source, sess.Status, core.FormatTime(sess.LastActivityAt),
		sess.Agent, sess.Tags, core.Truncate(sess.Title, 52))
}

func pinLabel(sess core.Session) string {
	if sess.Pinned {
		return "*"
	}
	return ""
}

func contentMatches(ctx context.Context, sessions []core.Session, query string) []core.Session {
	query = strings.ToLower(query)
	var out []core.Session
	for _, sess := range sessions {
		if sess.RawPath == "" {
			continue
		}
		items, err := adapters.ReadTextCandidates(sess.RawPath, 80)
		if err != nil {
			continue
		}
		for _, item := range items {
			if ctx.Err() != nil {
				return out
			}
			if strings.Contains(strings.ToLower(item), query) {
				if sess.GoalSummary == "" {
					sess.GoalSummary = core.Truncate(item, 180)
				}
				out = append(out, sess)
				break
			}
		}
	}
	return out
}

func runFZF(sessions []core.Session) (string, error) {
	if _, err := exec.LookPath("fzf"); err != nil {
		return "", fmt.Errorf("fzf not found in PATH; run without --fzf or install fzf")
	}
	var input strings.Builder
	for _, sess := range sessions {
		fmt.Fprintf(&input, "%s\t%s\t%s\t%s\t%s\n",
			sess.ID, sess.Source, core.ProjectName(sess.ProjectPath), sess.Title, sess.GoalSummary)
	}
	fzf := exec.Command("fzf", "--with-nth=1,2,3,4,5", "--delimiter=\t")
	fzf.Stdin = strings.NewReader(input.String())
	fzf.Stderr = os.Stderr
	out, err := fzf.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 130 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func validStatus(status string) bool {
	switch status {
	case core.StatusNeedsAttention, core.StatusWorking, core.StatusIdle, core.StatusDone, core.StatusStale, core.StatusUnknown:
		return true
	default:
		return false
	}
}

func sourcePath(cfg config.Config, source string) string {
	switch source {
	case "claude":
		return cfg.Sources.ClaudeProjects
	case "copilot":
		return cfg.Sources.CopilotSessionState
	default:
		return ""
	}
}

func printSession(sess core.Session) {
	fmt.Printf("ID: %s\n", sess.ID)
	fmt.Printf("Source: %s\n", sess.Source)
	fmt.Printf("Native ID: %s\n", sess.NativeID)
	fmt.Printf("Agent: %s\n", sess.Agent)
	fmt.Printf("Status: %s\n", sess.Status)
	fmt.Printf("Title: %s\n", sess.Title)
	if sess.TitleSource != "" {
		fmt.Printf("Title source: %s\n", sess.TitleSource)
	}
	if sess.GoalSummary != "" {
		fmt.Printf("Goal: %s\n", sess.GoalSummary)
	}
	if sess.Pinned {
		fmt.Println("Pinned: true")
	}
	if sess.Tags != "" {
		fmt.Printf("Tags: %s\n", sess.Tags)
	}
	fmt.Printf("Project: %s\n", firstNonEmpty(sess.ProjectPath, "-"))
	fmt.Printf("Last activity: %s\n", core.FormatTime(sess.LastActivityAt))
	if sess.AttachCommand != "" {
		fmt.Printf("Attach: %s\n", sess.AttachCommand)
	}
	if sess.ResumeCommand != "" {
		fmt.Printf("Resume: %s\n", sess.ResumeCommand)
	}
	if sess.RawPath != "" {
		fmt.Printf("Raw path: %s\n", sess.RawPath)
	}
}

func existsLabel(path string) string {
	if core.ExistingDir(path) || core.ExistingFile(path) {
		return "found"
	}
	return "missing"
}

func friendlySessionErr(ref string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("session %q not found; run tss list or tss scan", ref)
	}
	return err
}

func openCommand(backend string, sess core.Session) string {
	if backend == "tmux" || !toolExists("herdr") {
		name := "tss-" + core.SanitizeName(core.ProjectName(sess.ProjectPath)+"-"+sess.ID)
		return "tmux new-session -A -s " + core.ShellQuote(name) + " -c " + core.ShellQuote(sess.ProjectPath)
	}
	return "herdr workspace create --cwd " + core.ShellQuote(sess.ProjectPath) + " --label " + core.ShellQuote(sess.Title)
}

func sendCommand(rt core.RuntimeInstance, text string) (string, error) {
	switch rt.Backend {
	case "herdr":
		return "herdr agent prompt " + core.ShellQuote(rt.NativeID) + " " + core.ShellQuote(text), nil
	case "tmux":
		return "tmux send-keys -t " + core.ShellQuote(rt.NativeID) + " " + core.ShellQuote(text) + " Enter", nil
	default:
		return "", fmt.Errorf("backend %q cannot send text yet", rt.Backend)
	}
}

func readCommand(rt core.RuntimeInstance, lines int) (string, error) {
	if lines <= 0 {
		lines = 80
	}
	switch rt.Backend {
	case "herdr":
		return "herdr agent read " + core.ShellQuote(rt.NativeID) + " --source recent --lines " + fmt.Sprint(lines), nil
	case "tmux":
		return "tmux capture-pane -p -t " + core.ShellQuote(rt.NativeID) + " -S -" + fmt.Sprint(lines), nil
	default:
		return "", fmt.Errorf("backend %q cannot read output yet", rt.Backend)
	}
}

func runCommand(rt core.RuntimeInstance, sess core.Session, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}
	cwd := firstNonEmpty(rt.ProjectPath, sess.ProjectPath)
	switch rt.Backend {
	case "herdr":
		pane := firstNonEmpty(rt.Surface, rt.NativeID)
		if pane == "" {
			return "", fmt.Errorf("herdr runtime has no pane target")
		}
		split := "herdr pane split " + core.ShellQuote(pane) + " --direction right --no-focus"
		if cwd != "" {
			split += " --cwd " + core.ShellQuote(cwd)
		}
		return "pane=$(" + split + " | python3 -c 'import json,sys; print(json.load(sys.stdin)[\"result\"][\"pane\"][\"pane_id\"])')" +
			" && herdr pane run \"$pane\" " + core.ShellQuote(command), nil
	case "tmux":
		args := "tmux split-window -h -t " + core.ShellQuote(rt.NativeID)
		if cwd != "" {
			args += " -c " + core.ShellQuote(cwd)
		}
		return args + " " + core.ShellQuote(command), nil
	default:
		return "", fmt.Errorf("backend %q cannot run commands yet", rt.Backend)
	}
}

func runInteractive(command string) error {
	c := exec.Command("sh", "-lc", command)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func toolExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func init() {
	cobra.EnableCommandSorting = false
	_ = context.Background()
	_ = time.Now()
}
