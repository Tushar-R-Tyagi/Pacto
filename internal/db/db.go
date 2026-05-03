package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// ── Init ─────────────────────────────────────────────────

func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// Enable WAL mode for better concurrent access
	DB.Exec("PRAGMA journal_mode=WAL")
	DB.Exec("PRAGMA foreign_keys=ON")

	if err := createTables(); err != nil {
		return err
	}
	if err := insertDefaults(); err != nil {
		return err
	}
	return nil
}

func createTables() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id                     INTEGER PRIMARY KEY AUTOINCREMENT,
			planned_outcome        TEXT    NOT NULL,
			bout_minutes           INTEGER NOT NULL,
			status                 TEXT    NOT NULL DEFAULT 'started',
			shifted_to_session_id  INTEGER,
			created_at             TEXT    NOT NULL,
			completed_at           TEXT
		);
	`)
	return err
}

func insertDefaults() error {
	defaults := map[string]string{
		"outcome":        "",
		"bout_minutes":   "45",
		"theme":          "theme1",
		"shifted":        "0",
		"chimes_enabled": "1",
		"music_enabled":  "0",
		"music_track":    "",
		"music_volume":   "0.3",
		"chime_volume":   "0.7",
	}
	for key, val := range defaults {
		_, err := DB.Exec(
			"INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)",
			key, val,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// ── Settings ─────────────────────────────────────────────

func GetSetting(key string) string {
	var val string
	DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	return val
}

func SetSetting(key, value string) error {
	_, err := DB.Exec(
		"INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

// ── Sessions ─────────────────────────────────────────────

type Session struct {
	ID                 int64
	PlannedOutcome     string
	BoutMinutes        int
	Status             string
	ShiftedToSessionID *int64 // nullable
	CreatedAt          string
	CompletedAt        *string // nullable
}

func CreateSession(outcome string, minutes int) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := DB.Exec(
		"INSERT INTO sessions (planned_outcome, bout_minutes, status, created_at) VALUES (?, ?, 'started', ?)",
		outcome, minutes, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func CompleteSession(id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sessions SET status = 'completed', completed_at = ? WHERE id = ?",
		now, id,
	)
	return err
}

func ShiftSession(id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sessions SET status = 'shifted', completed_at = ? WHERE id = ?",
		now, id,
	)
	return err
}

func DiscardSession(id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"UPDATE sessions SET status = 'discarded', completed_at = ? WHERE id = ?",
		now, id,
	)
	return err
}

func LinkShiftedSession(oldID, newID int64) error {
	_, err := DB.Exec(
		"UPDATE sessions SET shifted_to_session_id = ? WHERE id = ?",
		newID, oldID,
	)
	return err
}

func GetRecentSessions(limit int) ([]Session, error) {
	rows, err := DB.Query(
		`SELECT id, planned_outcome, bout_minutes, status,
		        shifted_to_session_id, created_at, completed_at
		 FROM sessions ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		rows.Scan(
			&s.ID, &s.PlannedOutcome, &s.BoutMinutes, &s.Status,
			&s.ShiftedToSessionID, &s.CreatedAt, &s.CompletedAt,
		)
		sessions = append(sessions, s)
	}
	return sessions, nil
}

type Stats struct {
	TotalSessions       int
	Completed           int
	Shifted             int
	Discarded           int
	TotalMinutesPlanned int
	CompletionRate      float64
}

func GetStats() (Stats, error) {
	var s Stats
	row := DB.QueryRow("SELECT COUNT(*) FROM sessions")
	row.Scan(&s.TotalSessions)

	row = DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'completed'")
	row.Scan(&s.Completed)

	row = DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'shifted'")
	row.Scan(&s.Shifted)

	row = DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'discarded'")
	row.Scan(&s.Discarded)

	row = DB.QueryRow("SELECT COALESCE(SUM(bout_minutes), 0) FROM sessions")
	row.Scan(&s.TotalMinutesPlanned)

	if s.TotalSessions > 0 {
		s.CompletionRate = float64(s.Completed) / float64(s.TotalSessions) * 100
	}

	return s, nil
}

// Close the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}

type DaySessions struct {
	Date     string
	Sessions []Session
}

func GetSessionsByDate(limit int) ([]DaySessions, error) {
	rows, err := DB.Query(
		`SELECT id, planned_outcome, bout_minutes, status,
		        shifted_to_session_id, created_at, completed_at
		 FROM sessions 
		 ORDER BY created_at DESC 
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group by date
	days := make(map[string][]Session)
	var dayOrder []string

	for rows.Next() {
		var s Session
		rows.Scan(
			&s.ID, &s.PlannedOutcome, &s.BoutMinutes, &s.Status,
			&s.ShiftedToSessionID, &s.CreatedAt, &s.CompletedAt,
		)
		// Extract date part from created_at (YYYY-MM-DD)
		date := s.CreatedAt[:10]
		if _, exists := days[date]; !exists {
			dayOrder = append(dayOrder, date)
		}
		days[date] = append(days[date], s)
	}

	var result []DaySessions
	for _, date := range dayOrder {
		result = append(result, DaySessions{
			Date:     date,
			Sessions: days[date],
		})
	}

	return result, nil
}
