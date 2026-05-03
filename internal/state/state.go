package state

import (
	"pacto/internal/db"
	"strconv"
)

type AppState struct {
	outcome         string
	boutMinutes     int
	theme           string
	shifted         bool
	timerRunning    bool
	timeLeft        int
	activeSessionID int64

	chimesEnabled bool
	musicEnabled  bool
	musicTrack    string
	musicVolume   float64
	chimeVolume   float64

	listeners []func()
}

func New() *AppState {
	db.Init("pacto.db")

	s := &AppState{}
	s.outcome = db.GetSetting("outcome")
	s.boutMinutes, _ = strconv.Atoi(db.GetSetting("bout_minutes"))
	s.theme = db.GetSetting("theme")
	if s.theme == "" {
		s.theme = "theme1"
	}
	s.shifted = db.GetSetting("shifted") == "1"
	s.chimesEnabled = db.GetSetting("chimes_enabled") == "1"
	s.musicEnabled = db.GetSetting("music_enabled") == "1"
	s.musicTrack = db.GetSetting("music_track")
	s.musicVolume, _ = strconv.ParseFloat(db.GetSetting("music_volume"), 64)
	s.chimeVolume, _ = strconv.ParseFloat(db.GetSetting("chime_volume"), 64)
	return s
}

// ── Getters ──────────────────────────────────────────────

func (s *AppState) Outcome() string        { return s.outcome }
func (s *AppState) BoutMinutes() int       { return s.boutMinutes }
func (s *AppState) Theme() string          { return s.theme }
func (s *AppState) Shifted() bool          { return s.shifted }
func (s *AppState) TimerRunning() bool     { return s.timerRunning }
func (s *AppState) TimeLeft() int          { return s.timeLeft }
func (s *AppState) HasOutcome() bool       { return s.outcome != "" }
func (s *AppState) ActiveSessionID() int64 { return s.activeSessionID }
func (s *AppState) ChimesEnabled() bool    { return s.chimesEnabled }
func (s *AppState) MusicEnabled() bool     { return s.musicEnabled }
func (s *AppState) MusicTrack() string     { return s.musicTrack }
func (s *AppState) MusicVolume() float64   { return s.musicVolume }
func (s *AppState) ChimeVolume() float64   { return s.chimeVolume }

// ── Setters ──────────────────────────────────────────────

func (s *AppState) SetOutcome(text string) {
	s.outcome = text
	db.SetSetting("outcome", text)
	s.notify()
}

func (s *AppState) ClearOutcome() {
	s.outcome = ""
	s.shifted = false
	db.SetSetting("outcome", "")
	db.SetSetting("shifted", "0")
	s.notify()
}

func (s *AppState) ShiftOutcome() {
	s.shifted = true
	db.SetSetting("shifted", "1")
	db.SetSetting("outcome", s.outcome)
	s.notify()
}

func (s *AppState) UnshiftOutcome() {
	s.shifted = false
	db.SetSetting("shifted", "0")
	s.notify()
}

func (s *AppState) SetBoutMinutes(minutes int) {
	s.boutMinutes = minutes
	db.SetSetting("bout_minutes", strconv.Itoa(minutes))
	s.notify()
}

func (s *AppState) SetTheme(theme string) {
	s.theme = theme
	db.SetSetting("theme", theme)
	s.notify()
}

func (s *AppState) SetTimerRunning(running bool, timeLeft int) {
	s.timerRunning = running
	s.timeLeft = timeLeft
	s.notify()
}

func (s *AppState) UpdateTimeLeft(seconds int) {
	s.timeLeft = seconds
	s.notify()
}

// ── Audio setters ────────────────────────────────────────

func (s *AppState) SetChimesEnabled(enabled bool) {
	s.chimesEnabled = enabled
	if enabled {
		db.SetSetting("chimes_enabled", "1")
	} else {
		db.SetSetting("chimes_enabled", "0")
	}
	s.notify()
}

func (s *AppState) SetMusicEnabled(enabled bool) {
	s.musicEnabled = enabled
	if enabled {
		db.SetSetting("music_enabled", "1")
	} else {
		db.SetSetting("music_enabled", "0")
	}
	s.notify()
}

func (s *AppState) SetMusicTrack(track string) {
	s.musicTrack = track
	db.SetSetting("music_track", track)
	s.notify()
}

func (s *AppState) SetMusicVolume(vol float64) {
	s.musicVolume = vol
	db.SetSetting("music_volume", strconv.FormatFloat(vol, 'f', 2, 64))
	s.notify()
}

func (s *AppState) SetChimeVolume(vol float64) {
	s.chimeVolume = vol
	db.SetSetting("chime_volume", strconv.FormatFloat(vol, 'f', 2, 64))
	s.notify()
}

// ── Session logging ──────────────────────────────────────

func (s *AppState) StartSession() int64 {
	previousID := s.activeSessionID
	id, _ := db.CreateSession(s.outcome, s.boutMinutes)
	s.activeSessionID = id
	if s.shifted && previousID != 0 {
		db.LinkShiftedSession(previousID, id)
	}
	return id
}

func (s *AppState) CompleteSession() {
	if s.activeSessionID != 0 {
		db.CompleteSession(s.activeSessionID)
		s.activeSessionID = 0
	}
}

func (s *AppState) ShiftCurrentSession() {
	if s.activeSessionID != 0 {
		db.ShiftSession(s.activeSessionID)
	}
}

func (s *AppState) DiscardSession() {
	if s.activeSessionID != 0 {
		db.DiscardSession(s.activeSessionID)
		s.activeSessionID = 0
	}
}

// ── Observer pattern ─────────────────────────────────────

func (s *AppState) AddListener(callback func()) {
	s.listeners = append(s.listeners, callback)
}

func (s *AppState) notify() {
	for _, listener := range s.listeners {
		listener()
	}
}
