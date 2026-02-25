package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/zjom/pom/internal/pomodoro"
	"github.com/zjom/pom/internal/storage"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show aggregated session statistics",
	RunE:  runSummary,
}

var (
	sumName     string
	sumFrom     string
	sumTo       string
	sumType     string
	sumJSON     bool
)

func init() {
	summaryCmd.Flags().StringVar(&sumName, "name", "", "filter by session name")
	summaryCmd.Flags().StringVar(&sumFrom, "from", "", "start date (YYYY-MM-DD)")
	summaryCmd.Flags().StringVar(&sumTo, "to", "", "end date (YYYY-MM-DD)")
	summaryCmd.Flags().StringVar(&sumType, "type", "", "filter by type: focus, short-break, long-break")
	summaryCmd.Flags().BoolVar(&sumJSON, "json", false, "output as JSON")

	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("determine database path: %w", err)
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	f, err := buildFilter(sumName, sumFrom, sumTo, sumType, 0)
	if err != nil {
		return err
	}

	stats, err := store.GetStatistics(context.Background(), f)
	if err != nil {
		return fmt.Errorf("query statistics: %w", err)
	}

	if sumJSON {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if stats.TotalSessions == 0 {
		fmt.Println("No sessions found matching the given filters.")
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	fmt.Println(headerStyle.Render("📊 Session Summary"))
	fmt.Println()

	rows := [][]string{
		{"Total Sessions", fmt.Sprintf("%d", stats.TotalSessions)},
		{"Total Time", formatDuration(stats.TotalTime)},
		{"Average Duration", formatDuration(stats.AverageDuration)},
	}

	for typ, count := range stats.ByType {
		rows = append(rows, []string{typ, fmt.Sprintf("%d", count)})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("Metric", "Value").
		Rows(rows...)

	fmt.Println(t)

	return nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func buildFilter(name, from, to, typ string, limit int) (storage.QueryFilter, error) {
	f := storage.QueryFilter{
		Name:  name,
		Limit: limit,
	}

	if from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			return f, fmt.Errorf("invalid --from date: %w", err)
		}
		f.From = &t
	}

	if to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			return f, fmt.Errorf("invalid --to date: %w", err)
		}
		// Include the full day.
		end := t.Add(24*time.Hour - time.Second)
		f.To = &end
	}

	if typ != "" {
		var st pomodoro.SessionType
		switch typ {
		case "focus":
			st = pomodoro.Focus
		case "short-break":
			st = pomodoro.ShortBreak
		case "long-break":
			st = pomodoro.LongBreak
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown type %q, ignoring filter\n", typ)
			return f, nil
		}
		f.SessionType = &st
	}

	return f, nil
}
