package db

import (
	"database/sql"
)

func LogSession(db *sql.DB, duration, reviewed, finished int) error {
	_, err := db.Exec("INSERT INTO sessions (duration, reviewed, finished) VALUES (?, ?, ?)", duration, reviewed, finished)
	return err
}

type SessionStats struct {
	Date     string
	Duration int
	Reviewed int
	Finished int
	AvgPer   float64
}

func GetRecentSessions(db *sql.DB, limit int) ([]SessionStats, error) {
	q := `SELECT date, SUM(duration), SUM(reviewed), SUM(finished)
	      FROM sessions GROUP BY date ORDER BY date DESC LIMIT ?`

	rows, err := db.Query(q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []SessionStats
	for rows.Next() {
		var s SessionStats
		var totalDuration, totalReviewed int
		if err := rows.Scan(&s.Date, &totalDuration, &totalReviewed, &s.Finished); err != nil {
			return nil, err
		}
		s.Duration = totalDuration
		s.Reviewed = totalReviewed
		if s.Reviewed > 0 {
			s.AvgPer = float64(totalDuration) / float64(totalReviewed) / 60.0
		}
		stats = append(stats, s)
	}
	return stats, nil
}
