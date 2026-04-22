package commands

import (
	"database/sql"
	"fmt"

	"github.com/psto/irw/internal/db"
)

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
