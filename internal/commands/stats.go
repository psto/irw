package commands

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/tui"
)

func ShowStats(database *sql.DB, trackType string) {
	if trackType == "" {
		trackType = "reading"
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

func ReadFile(database *sql.DB) error {
	paths, err := db.GetActivePaths(database)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		fmt.Println("No files to read.")
		return nil
	}

	selected := paths[0]

	startTime := time.Now()

	if err := tui.LaunchFile(selected); err != nil {
		return err
	}

	fmt.Println("Press Enter when done reading...")
	tui.ReadKey()

	duration := int(time.Since(startTime).Seconds())
	db.LogSession(database, duration, 1, 0)

	fmt.Printf("Read: %s (%dm %ds)\n", tui.BaseName(selected), duration/60, duration%60)
	return nil
}

func PurgeFinished(database *sql.DB) error {
	count, err := db.CountFinished(database)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d finished items.\n", count)

	confirm, err := Confirm("Delete them from database?")
	if err != nil {
		return err
	}

	if confirm {
		if err := db.PurgeFinished(database); err != nil {
			return err
		}
		fmt.Println("Purged.")
	}
	return nil
}
