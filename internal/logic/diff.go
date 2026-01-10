package logic

import (
	"encoding/json"

	"github.com/navyaalva/sbf-os/internal/db"
)

type Change struct {
	Field string      `json:"field"`
	From  interface{} `json:"from"`
	To    interface{} `json:"to"`
}

func CalculateChanges(oldT, newT db.Task) []byte {
	var changes []Change

	if oldT.Title != newT.Title {
		changes = append(changes, Change{Field: "title", From: oldT.Title, To: newT.Title})
	}
	if oldT.Description.String != newT.Description.String || oldT.Description.Valid != newT.Description.Valid {
		changes = append(changes, Change{Field: "description", From: oldT.Description, To: newT.Description})
	}
	if oldT.Status != newT.Status {
		changes = append(changes, Change{Field: "status", From: oldT.Status, To: newT.Status})
	}
	if oldT.Priority != newT.Priority {
		changes = append(changes, Change{Field: "priority", From: oldT.Priority, To: newT.Priority})
	}

	// due_date, tags if you want (optional)

	if len(changes) == 0 {
		return []byte(`[]`)
	}

	b, _ := json.Marshal(changes)
	return b
}
