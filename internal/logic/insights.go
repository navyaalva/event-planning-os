package logic

import (
	"time"

	"github.com/navyaalva/sbf-os/internal/db"
)

type ScoredTask struct {
	Task      interface{} // Can hold db.GetEventTasksRow or db.GetGlobalActiveTasksRow
	Score     int
	Reasons   []string
	RiskLevel string
}

// Common scoring engine
func calculateRisk(status string, priority int32, dueDate, createdAt, lastUpdate time.Time, isDueDateValid, isLastUpdateValid bool) (int, []string, string) {
	score := 0
	reasons := []string{}
	now := time.Now()

	// 1. Due Date
	if isDueDateValid {
		daysUntil := int(time.Until(dueDate).Hours() / 24)
		if daysUntil < 0 {
			score += 50
			reasons = append(reasons, "OVERDUE")
		} else if daysUntil <= 3 {
			score += 30
			reasons = append(reasons, "Due Soon")
		}
	}

	// 2. Staleness
	lastTouch := createdAt
	if isLastUpdateValid {
		lastTouch = lastUpdate
	}
	daysSince := int(now.Sub(lastTouch).Hours() / 24)
	if status != "done" {
		if daysSince >= 14 {
			score += 30
			reasons = append(reasons, "Stale (14d)")
		} else if daysSince >= 7 {
			score += 10
			reasons = append(reasons, "Stale (7d)")
		}
	}

	// 3. Status
	if status == "blocked" {
		score += 25
		reasons = append(reasons, "Blocked")
	}

	// 4. Priority
	score += int(priority) * 5

	// Level
	level := "low"
	if score >= 50 {
		level = "high"
	} else if score >= 25 {
		level = "med"
	}

	return score, reasons, level
}

// Wrapper for Event View
func ScoreTaskRow(t db.GetEventTasksRow) ScoredTask {
	// Helper to convert sql.NullTime to generic inputs
	due := time.Time{}
	if t.DueDate.Valid {
		due = t.DueDate.Time
	}
	upd := time.Time{}
	if t.LastUpdateAt.Valid {
		upd = t.LastUpdateAt.Time
	}

	s, r, l := calculateRisk(t.Status, t.Priority, due, t.CreatedAt, upd, t.DueDate.Valid, t.LastUpdateAt.Valid)
	return ScoredTask{Task: t, Score: s, Reasons: r, RiskLevel: l}
}

// Wrapper for Global View (AI Briefing)
func ScoreGlobalTask(t db.GetGlobalActiveTasksRow) ScoredTask {
	due := time.Time{}
	if t.DueDate.Valid {
		due = t.DueDate.Time
	}
	upd := time.Time{}
	if t.LastUpdateAt.Valid {
		upd = t.LastUpdateAt.Time
	}

	s, r, l := calculateRisk(t.Status, t.Priority, due, t.CreatedAt, upd, t.DueDate.Valid, t.LastUpdateAt.Valid)
	return ScoredTask{Task: t, Score: s, Reasons: r, RiskLevel: l}
}
