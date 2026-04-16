package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/models"
	"github.com/psto/irw/internal/tui"
)

func Import(database *sql.DB) {
	importFromZk(database)
	importFromSioyek(database)
}

func importFromZk(database *sql.DB) {
	for _, tag := range []string{"reading", "writing"} {
		fmt.Printf("Importing status/%s...\n", tag)
		output, err := tui.RunZk("list", "--tag", "status/"+tag, "--format", "json")
		if err != nil {
			continue
		}

		var notes []models.ZkNote
		if err := json.Unmarshal(output, &notes); err != nil {
			continue
		}

		for _, note := range notes {
			priority := note.Metadata.Priority
			if err := db.InsertTrack(database, note.AbsPath, tag, priority); err != nil {
				continue
			}
			if priority != nil {
				db.UpdatePriority(database, note.AbsPath, *priority)
			}
		}
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
