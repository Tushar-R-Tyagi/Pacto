package main

import (
	"context"
	"pacto/internal/audio"
	"pacto/internal/db"
	"pacto/internal/state"
	"pacto/internal/timer"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx   context.Context
	state *state.AppState
	timer *timer.BoutTimer
}

func NewApp(s *state.AppState) *App {
	return &App{
		state: s,
		timer: timer.New(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ── Frontend calls these ─────────────────────────────────

func (a *App) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"outcome":      a.state.Outcome(),
		"boutMinutes":  a.state.BoutMinutes(),
		"theme":        a.state.Theme(),
		"hasOutcome":   a.state.HasOutcome(),
		"isShifted":    a.state.Shifted(),
		"timerRunning": a.timer.Running,
		"timeLeft":     a.timer.SecondsLeft,
	}
}

func (a *App) SetTheme(theme string) {
	a.state.SetTheme(theme)
}

func (a *App) SetOutcome(text string) {
	a.state.SetOutcome(text)
}

func (a *App) SetBoutMinutes(minutes int) {
	a.state.SetBoutMinutes(minutes)
}

func (a *App) StartBout() string {
	if a.timer.Running {
		return "already running"
	}
	a.state.UnshiftOutcome()
	a.state.StartSession()
	a.timer.Start(a.state.BoutMinutes() * 60)
	if a.state.ChimesEnabled() {
		audio.PlayStartChime()
	}

	//Minimise window when timer starts
	runtime.WindowMinimise(a.ctx)

	go func() {
		<-a.timer.Done
		if a.state.ChimesEnabled() {
			audio.PlayEndChime()
		}
		// Show window when timer ends
		runtime.WindowShow(a.ctx)
		runtime.WindowUnminimise(a.ctx)
	}()
	return "ok"
}

func (a *App) CancelBout() string {
	a.timer.Stop()
	a.state.DiscardSession()
	a.state.ClearOutcome()
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	return "ok"
}

func (a *App) Review(result string) string {
	switch result {
	case "complete":
		a.state.CompleteSession()
		a.state.ClearOutcome()
	case "shift":
		a.state.ShiftCurrentSession()
		a.state.ShiftOutcome()
	case "discard":
		a.state.DiscardSession()
		a.state.ClearOutcome()
	}
	return "ok"
}

func (a *App) GetSessions() []map[string]interface{} {
	sessions, _ := db.GetRecentSessions(50)
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":             s.ID,
			"plannedOutcome": s.PlannedOutcome,
			"boutMinutes":    s.BoutMinutes,
			"status":         s.Status,
		})
	}
	return result
}

func (a *App) GetStats() map[string]interface{} {
	stats, _ := db.GetStats()
	return map[string]interface{}{
		"totalSessions": stats.TotalSessions,
		"completed":     stats.Completed,
		"shifted":       stats.Shifted,
		"discarded":     stats.Discarded,
	}
}
func (a *App) GetLog() []map[string]interface{} {
	days, _ := db.GetSessionsByDate(100)
	var result []map[string]interface{}
	for _, day := range days {
		var sessions []map[string]interface{}
		totalMinutes := 0
		for _, s := range day.Sessions {
			sessions = append(sessions, map[string]interface{}{
				"id":             s.ID,
				"plannedOutcome": s.PlannedOutcome,
				"boutMinutes":    s.BoutMinutes,
				"status":         s.Status,
				"createdAt":      s.CreatedAt,
				"completedAt":    s.CompletedAt,
			})
			if s.Status == "completed" {
				totalMinutes += s.BoutMinutes
			}
		}
		result = append(result, map[string]interface{}{
			"date":         day.Date,
			"sessions":     sessions,
			"totalMinutes": totalMinutes,
		})
	}
	return result
}
