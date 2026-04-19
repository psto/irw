package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/models"
	"github.com/psto/irw/internal/tui"
)

func Import(database *sql.DB, cfg config.ConfigProvider) {
	importFromZk(database, cfg, tui.RealZkRunner{})
	importFromSioyek(database)
}

func importFromZk(database *sql.DB, cfg config.ConfigProvider, zkRunner tui.ZkRunner) {
	zkTags := cfg.GetZkTags()
	validPaths := make(map[string]bool)

	for trackType, tags := range zkTags {
		for _, tag := range tags {
			fmt.Printf("Importing %s...\n", tag)
			output, err := zkRunner.Run("list", "--tag", tag, "--format", "json")
			if err != nil {
				continue
			}

			var notes []models.ZkNote
			if err := json.Unmarshal(output, &notes); err != nil {
				continue
			}

			for _, note := range notes {
				validPaths[note.AbsPath] = true
				priority := note.Metadata.Priority
				if err := db.InsertTrack(database, note.AbsPath, trackType, priority); err != nil {
					continue
				}
				if priority != nil {
					db.UpdatePriority(database, note.AbsPath, *priority)
				}
			}
		}
	}

	configuredTypes := make([]string, 0, len(zkTags))
	for t := range zkTags {
		configuredTypes = append(configuredTypes, t)
	}
	if err := db.CleanupOrphanedTracks(database, configuredTypes, validPaths); err != nil {
		fmt.Printf("Warning: failed to cleanup orphaned tracks: %v\n", err)
	}

	fmt.Println("Checking for dead links...")
	paths, _ := db.GetAllPaths(database)
	for _, storedPath := range paths {
		if isURI(storedPath) {
			continue
		}
		if !tui.FileExists(storedPath) {
			fmt.Printf("Removing dead file: %s\n", storedPath)
			db.DeleteTrack(database, storedPath)
		}
	}
}

func importFromSioyek(database *sql.DB) {
	fmt.Println("Syncing highlighted files from Sioyek to tracker...")

	home, _ := tui.GetHomeDir()
	sioyekSharedDB := filepath.Join(home, ".local/share/sioyek/shared.db")
	sioyekLocalDB := filepath.Join(home, ".local/share/sioyek/local.db")

	sharedDB, err := sql.Open("sqlite", sioyekSharedDB)
	if err != nil {
		return
	}
	defer sharedDB.Close()

	rows, err := sharedDB.Query(`SELECT DISTINCT document_path FROM highlights 
		WHERE creation_time > datetime('now', '-1 day')`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var docPath string
		if err := rows.Scan(&docPath); err != nil {
			continue
		}

		realPath := docPath
		if !strings.Contains(docPath, "/") {
			localDB, err := sql.Open("sqlite", sioyekLocalDB)
			if err != nil {
				continue
			}
			localDB.QueryRow("SELECT path FROM document_hash WHERE hash=? LIMIT 1", docPath).Scan(&realPath)
			localDB.Close()
		}

		if tui.FileExists(realPath) {
			exists, _ := db.PathExists(database, realPath)
			if !exists {
				db.InsertTrack(database, realPath, "reading", nil)
				fmt.Printf("Added to tracker from Sioyek highlights: %s\n", tui.BaseName(realPath))
			}
		}
	}
}
