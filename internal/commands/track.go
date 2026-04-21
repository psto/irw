package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/models"
	"github.com/psto/irw/internal/tui"
)

func Track(cfg config.ConfigProvider, database *sql.DB, input string, queueType string) error {
	if queueType == "" {
		queueType = cfg.GetDefaultQueue()
	}

	validQueues := cfg.GetZkTags()
	if _, ok := validQueues[queueType]; !ok {
		queues := make([]string, 0, len(validQueues))
		for q := range validQueues {
			queues = append(queues, q)
		}
		sort.Strings(queues)
		return fmt.Errorf("invalid queue: %q (configured queues: %s)", queueType, strings.Join(queues, ", "))
	}

	if isURI(input) {
		if err := db.InsertTrack(database, input, queueType, nil); err != nil {
			return err
		}
		fmt.Printf("Tracked URI: %s\n", input)
		return nil
	}

	if !tui.FileExists(input) {
		return fmt.Errorf("file not found: %s", input)
	}

	absPath, err := tui.AbsPath(input)
	if err != nil {
		return err
	}

	var priority *int
	if strings.HasSuffix(absPath, ".md") {
		priority = extractZkPriority(absPath)
	}

	if err := db.InsertTrack(database, absPath, queueType, priority); err != nil {
		return err
	}

	if priority != nil {
		fmt.Printf("Tracked (MD Priority %d): %s\n", *priority, tui.BaseName(absPath))
	} else {
		fmt.Printf("Tracked: %s\n", tui.BaseName(absPath))
	}
	return nil
}

func isURI(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "zotero://")
}

func extractZkPriority(path string) *int {
	output, err := tui.RunZk("list", path, "--format", "json")
	if err != nil {
		return nil
	}

	var notes []models.ZkNote
	if err := json.Unmarshal(output, &notes); err != nil || len(notes) == 0 {
		return nil
	}

	if notes[0].Metadata.Priority != nil {
		return notes[0].Metadata.Priority
	}
	return nil
}

func Untrack(database *sql.DB, target string) error {
	if target == "" {
		return fmt.Errorf("file argument required")
	}

	if err := db.DeleteTrack(database, target); err != nil {
		return err
	}
	fmt.Printf("Untracked: %s\n", target)
	return nil
}

func Complete(database *sql.DB, target string) error {
	if target == "" {
		return fmt.Errorf("file argument required")
	}

	if err := db.MarkFinished(database, target); err != nil {
		return err
	}
	fmt.Printf("Completed: %s\n", target)
	return nil
}

func SetPriority(database *sql.DB, target string, newPriority int) error {
	if target == "" {
		return fmt.Errorf("file argument required")
	}

	if err := db.UpdatePriority(database, target, newPriority); err != nil {
		return err
	}
	fmt.Printf("Priority %d set for %s\n", newPriority, target)
	return nil
}
