package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/tui"
)

func ShowStats(cfg config.ConfigProvider, database *sql.DB, trackType string) {
	if trackType == "" {
		trackType = cfg.GetDefaultQueue()
	}

	active, finished, due, completion, err := db.GetTrackStats(database, trackType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting stats: %v\n", err)
		return
	}

	fmt.Print(tui.RenderStats(trackType, active, finished, due, completion))

	sessions, err := db.GetRecentSessions(database, 7)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sessions: %v\n", err)
		return
	}

	fmt.Print(tui.RenderSessions(sessions))
}

func ShowSchedule(database *sql.DB, raw, null bool) {
	items, err := db.GetSchedule(database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting schedule: %v\n", err)
		return
	}

	if null {
		for _, item := range items {
			fmt.Printf("%s\x00", item.Path)
		}
		return
	}

	if raw {
		for _, item := range items {
			fmt.Printf("%s,%.1f,%.1f,%d,%s,%s\n", item.DueDate, item.Interval, item.Afactor, item.Priority, item.Type, item.Path)
		}
		return
	}

	fmt.Print(tui.RenderSchedule(items))
}
