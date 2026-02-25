package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/zjom/pom/internal/storage"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "List completed sessions",
	RunE:  runHistory,
}

var (
	histName  string
	histFrom  string
	histTo    string
	histType  string
	histLimit int
	histJSON  bool
)

func init() {
	historyCmd.Flags().StringVar(&histName, "name", "", "filter by session name")
	historyCmd.Flags().StringVar(&histFrom, "from", "", "start date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&histTo, "to", "", "end date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&histType, "type", "", "filter by type: focus, short-break, long-break")
	historyCmd.Flags().IntVar(&histLimit, "limit", 0, "max number of sessions to show")
	historyCmd.Flags().BoolVar(&histJSON, "json", false, "output as JSON")

	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("determine database path: %w", err)
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	f, err := buildFilter(histName, histFrom, histTo, histType, histLimit)
	if err != nil {
		return err
	}

	sessions, err := store.ListSessions(context.Background(), f)
	if err != nil {
		return fmt.Errorf("query sessions: %w", err)
	}

	if histJSON {
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found matching the given filters.")
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	fmt.Println(headerStyle.Render("📋 Session History"))
	fmt.Println()

	var rows [][]string
	for _, s := range sessions {
		rows = append(rows, []string{
			s.StartedAt.Local().Format("2006-01-02 15:04"),
			s.Name,
			string(s.SessionType),
			formatDuration(s.CompletedAt.Sub(s.StartedAt)),
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("Date", "Name", "Type", "Duration").
		Rows(rows...)

	fmt.Println(t)

	return nil
}
