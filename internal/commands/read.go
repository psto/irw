package commands

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/tui"
)

func ReadFile(cfg config.ConfigProvider, database *sql.DB, filePath string) error {
	var selected string

	if filePath != "" {
		absPath, err := tui.AbsPath(filePath)
		if err != nil {
			return err
		}
		exists, err := db.PathExists(database, absPath)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("file not tracked: %s\nRun: irw track %s", absPath, absPath)
		}
		selected = absPath
	} else {
		paths, err := db.GetActivePaths(database)
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			fmt.Println("No files to read.")
			return nil
		}
		selected, err = SelectFromList(paths, "Select a file to read")
		if err != nil {
			return err
		}
		if selected == "" {
			return nil
		}
	}

	startTime := time.Now()

	if err := tui.LaunchFile(cfg, selected); err != nil {
		return err
	}

	fmt.Println("Press Enter when done reading...")
	tui.ReadKey()

	duration := int(time.Since(startTime).Seconds())
	db.LogSession(database, duration, 1, 0)

	fmt.Printf("Read: %s (%dm %ds)\n", tui.BaseName(selected), duration/60, duration%60)
	return nil
}
