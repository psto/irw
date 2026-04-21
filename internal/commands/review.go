package commands

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/tui"
)

func Review(cfg config.ConfigProvider, database *sql.DB, trackType string, ext string, compactMode bool) error {
	if trackType == "" {
		trackType = cfg.GetDefaultQueue()
	}

	startTime := time.Now()
	var count, finished int

	for {
		dueCount, err := db.CountDue(database, trackType, ext)
		if err != nil {
			return err
		}

		path, interval, afactor, err := db.GetNextDue(database, trackType, ext)
		if err != nil {
			return err
		}
		if path == "" {
			fmt.Print(tui.RenderQueueEmpty())
			break
		}

		filename := filepath.Base(path)

		if compactMode {
			fmt.Print(tui.RenderReviewItemCompact(filename, dueCount))
		} else {
			fmt.Print(tui.RenderReviewItem(filename, dueCount))
		}

		if err := tui.LaunchFile(cfg, path); err != nil {
			fmt.Printf("Failed to launch: %v\n", err)
		}

		key, err := tui.ReadKey()
		if err != nil {
			return err
		}

		switch key {
		case 'p':
			input, err := Input("Priority (0-100): ", "")
			if err != nil {
				continue
			}
			var np int
			fmt.Sscanf(input, "%d", &np)
			db.UpdatePriority(database, path, np)

		case 'f':
			db.MarkFinished(database, path)
			finished++
			count++
			fmt.Println("Marked as finished.")

		case 's':
			db.SetDueIn(database, path, 3600)
			fmt.Println("Skipped for 1 hour.")

		case 'z':
			db.SetDueIn(database, path, 604800)
			fmt.Println("Postponed for 1 week.")

		case 'q', 3:
			fmt.Println("Quitting...")
			goto done

		case '\n', '\r':
			newDue := time.Now().Unix() + int64(interval*86400)
			newInterval := interval * afactor
			db.UpdateInterval(database, path, newInterval, newDue)
			count++
			fmt.Println("Rescheduled.")
		}
	}

done:
	duration := int(time.Since(startTime).Seconds())
	db.LogSession(database, duration, count, finished)

	ShowStats(cfg, database, trackType)
	return nil
}
