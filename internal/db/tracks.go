package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func InsertTrack(db *sql.DB, path string, trackType string, priority *int) error {
	var q string
	var args []any

	if priority != nil {
		q = "INSERT OR IGNORE INTO tracks (path, type, priority) VALUES (?, ?, ?)"
		args = []any{path, trackType, *priority}
	} else {
		q = "INSERT OR IGNORE INTO tracks (path, type) VALUES (?, ?)"
		args = []any{path, trackType}
	}

	_, err := db.Exec(q, args...)
	return err
}

func DeleteTrack(db *sql.DB, path string) error {
	_, err := db.Exec("DELETE FROM tracks WHERE path = ?", path)
	return err
}

func MarkFinished(db *sql.DB, path string) error {
	_, err := db.Exec("UPDATE tracks SET is_finished=1 WHERE path=?", path)
	return err
}

func UpdatePriority(db *sql.DB, path string, priority int) error {
	_, err := db.Exec("UPDATE tracks SET priority=? WHERE path=?", priority, path)
	return err
}

func UpdateInterval(db *sql.DB, path string, interval float64, dueDate int64) error {
	_, err := db.Exec("UPDATE tracks SET interval=?, due_date=? WHERE path=?", interval, dueDate, path)
	return err
}

func GetAllPaths(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT path FROM tracks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func GetNextDue(db *sql.DB, trackType string, ext string) (string, float64, float64, error) {
	q := `SELECT path, interval, afactor FROM tracks
	      WHERE due_date <= strftime('%s', 'now') AND is_finished=0`
	args := []any{}

	if trackType != "" {
		q += " AND type=?"
		args = append(args, trackType)
	}
	if ext != "" {
		q += " AND path LIKE ?"
		args = append(args, "%."+ext)
	}
	q += " ORDER BY (ABS(random() % 20) = 0) DESC, priority ASC, due_date ASC LIMIT 1"

	var path string
	var interval, afactor float64
	err := db.QueryRow(q, args...).Scan(&path, &interval, &afactor)
	if err == sql.ErrNoRows {
		return "", 0, 0, nil
	}
	return path, interval, afactor, err
}

func GetActivePaths(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT path FROM tracks WHERE is_finished = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func CountDue(db *sql.DB, trackType, ext string) (int, error) {
	q := `SELECT COUNT(*) FROM tracks WHERE due_date <= strftime('%s', 'now') AND is_finished=0`
	args := []any{}

	if trackType != "" {
		q += " AND type=?"
		args = append(args, trackType)
	}
	if ext != "" {
		q += " AND path LIKE ?"
		args = append(args, "%."+ext)
	}

	var count int
	err := db.QueryRow(q, args...).Scan(&count)
	return count, err
}

func GetTrackStats(db *sql.DB, trackType string) (active, finished, due int, completion float64, err error) {
	q := `SELECT 
		COUNT(*) FILTER (WHERE is_finished = 0) AS active,
		COUNT(*) FILTER (WHERE is_finished = 1) AS finished,
		COUNT(*) FILTER (WHERE is_finished = 0 AND due_date <= strftime('%s', 'now')) AS due
		FROM tracks WHERE type = ?`

	err = db.QueryRow(q, trackType).Scan(&active, &finished, &due)
	if err != nil {
		return
	}

	total := active + finished
	if total > 0 {
		completion = float64(finished) / float64(total) * 100
	}
	return
}

type ScheduleItem struct {
	DueDate  string
	Interval float64
	Afactor  float64
	Priority int
	Type     string
	Path     string
}

func GetSchedule(db *sql.DB) ([]ScheduleItem, error) {
	q := `SELECT datetime(due_date, 'unixepoch', 'localtime'), interval, afactor, priority, type, path 
	      FROM tracks WHERE is_finished = 0 ORDER BY due_date ASC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ScheduleItem
	for rows.Next() {
		var item ScheduleItem
		if err := rows.Scan(&item.DueDate, &item.Interval, &item.Afactor, &item.Priority, &item.Type, &item.Path); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func CountFinished(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tracks WHERE is_finished = 1").Scan(&count)
	return count, err
}

func PurgeFinished(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM tracks WHERE is_finished = 1")
	return err
}

func PathExists(db *sql.DB, path string) (bool, error) {
	var exists int
	err := db.QueryRow("SELECT 1 FROM tracks WHERE path=? LIMIT 1", path).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func SetDueIn(db *sql.DB, path string, seconds int64) error {
	newDue := time.Now().Unix() + seconds
	_, err := db.Exec("UPDATE tracks SET due_date=? WHERE path=?", newDue, path)
	return err
}

func CleanupOrphanedTracks(db *sql.DB, configuredTypes []string, validPaths map[string]bool) error {
	if len(configuredTypes) == 0 {
		return nil
	}

	if len(validPaths) == 0 {
		placeholders := strings.Repeat("?,", len(configuredTypes))
		placeholders = placeholders[:len(placeholders)-1]
		args := make([]any, len(configuredTypes))
		for i, t := range configuredTypes {
			args[i] = t
		}
		q := fmt.Sprintf("DELETE FROM tracks WHERE type IN (%s)", placeholders)
		_, err := db.Exec(q, args...)
		return err
	}

	placeholders := strings.Repeat("?,", len(configuredTypes))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(configuredTypes))
	for i, t := range configuredTypes {
		args[i] = t
	}

	q := fmt.Sprintf("SELECT path FROM tracks WHERE type IN (%s)", placeholders)
	rows, err := db.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var toDelete []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		if !validPaths[path] {
			toDelete = append(toDelete, path)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, path := range toDelete {
		if _, err := db.Exec("DELETE FROM tracks WHERE path = ?", path); err != nil {
			return err
		}
	}
	return nil
}
