package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	"github.com/navyaalva/sbf-os/internal/db"
	"github.com/navyaalva/sbf-os/internal/logic"
)

// Helper for display
type EventDisplay struct {
	ID             string
	Name           string
	Info           string
	Countdown      string
	CompletedTasks int64
	TotalTasks     int64
}

// 1) DASHBOARD
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	events, err := s.Q.ListEvents(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	displayEvents := make([]EventDisplay, 0, len(events))
	now := time.Now()

	for _, e := range events {
		daysLeft := int(math.Ceil(e.EventDate.Sub(now).Hours() / 24))
		countdown := fmt.Sprintf("%d days left", daysLeft)
		if daysLeft < 0 {
			countdown = "Event passed"
		} else if daysLeft == 0 {
			countdown = "Today!"
		}

		info := e.EventDate.Format("Jan 02, 2006")
		if e.Location.Valid && e.Location.String != "" {
			info += " • " + e.Location.String
		}

		displayEvents = append(displayEvents, EventDisplay{
			ID:             e.ID.String(),
			Name:           e.Name,
			Info:           info,
			Countdown:      countdown,
			CompletedTasks: e.CompletedTasks,
			TotalTasks:     e.TotalTasks,
		})
	}

	var briefingHTML template.HTML
	if r.URL.Query().Get("briefing") == "true" {
		tasks, err := s.Q.GetTasksForFollowUp(r.Context())
		if err == nil {
			reminders := logic.CheckFollowUps(tasks)
			if len(reminders) > 0 {
				var sb strings.Builder
				sb.WriteString("<strong>🔔 Suggested Follow-ups:</strong><ul>")
				for _, rem := range reminders {
					sb.WriteString("<li>")
					sb.WriteString(rem)
					sb.WriteString("</li>")
				}
				sb.WriteString("</ul>")
				briefingHTML = template.HTML(sb.String())
			} else {
				briefingHTML = template.HTML("✅ No urgent follow-ups needed.")
			}
		}
	}

	data := struct {
		Events   []EventDisplay
		Briefing template.HTML
	}{
		Events:   displayEvents,
		Briefing: briefingHTML,
	}

	tmpl := s.templates["dashboard"]
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// 2) EVENT DETAIL
func (s *Server) handleEventDetail(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	showAll := r.URL.Query().Get("show_all") == "on"

	tasks, err := s.Q.GetEventTasks(r.Context(), db.GetEventTasksParams{
		EventID: eventID,
		Column2: showAll,
	})
	if err != nil {
		http.Error(w, "Failed to fetch tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	grouped := make(map[string][]logic.ScoredTask)
	for _, t := range tasks {
		scored := logic.ScoreTaskRow(t)
		grouped[t.Category] = append(grouped[t.Category], scored)
	}

	data := struct {
		EventName       string
		EventID         string
		TasksByCategory map[string][]logic.ScoredTask
		ShowAll         bool
	}{
		EventName:       "Event Tasks",
		EventID:         eventID.String(),
		TasksByCategory: grouped,
		ShowAll:         showAll,
	}

	event, err := s.Q.GetEvent(r.Context(), eventID)
	if err == nil {
		data.EventName = event.Name
	}

	tmpl := s.templates["list_tasks"]
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// 3) CREATE TASK
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	prefillEventID := r.URL.Query().Get("event_id")

	switch r.Method {
	case http.MethodGet:
		events, _ := s.Q.ListEvents(r.Context())
		people, _ := s.Q.ListPeople(r.Context())

		tmpl := s.templates["create_task"]
		if tmpl == nil {
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}

		data := struct {
			EventID string
			Events  []db.ListEventsRow
			People  []db.Person
		}{
			EventID: prefillEventID,
			Events:  events,
			People:  people,
		}
		tmpl.ExecuteTemplate(w, "base", data)
		return

	case http.MethodPost:
		// Logic below
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	category := r.FormValue("category")
	eventIDStr := r.FormValue("event_id")
	ownerIDStr := r.FormValue("owner_id")
	assigneeName := r.FormValue("assignee_text")
	priorityInt, _ := strconv.Atoi(r.FormValue("priority"))
	if priorityInt == 0 {
		priorityInt = 3
	}
	if category == "" {
		category = "general"
	}
	descRaw := r.FormValue("description")
	dateRaw := r.FormValue("due_date")

	var assigneeParam sql.NullString
	if assigneeName != "" {
		assigneeParam = sql.NullString{String: assigneeName, Valid: true}
	}
	var descParam sql.NullString
	if descRaw != "" {
		descParam = sql.NullString{String: descRaw, Valid: true}
	}
	var dateParam sql.NullTime
	if dateRaw != "" {
		if t, err := time.Parse("2006-01-02", dateRaw); err == nil {
			dateParam = sql.NullTime{Time: t, Valid: true}
		}
	}
	var ownerIDParam uuid.NullUUID
	if ownerIDStr != "" {
		if p, err := uuid.Parse(ownerIDStr); err == nil {
			ownerIDParam = uuid.NullUUID{UUID: p, Valid: true}
		}
	}
	eventUUID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Error: You must select an Event.", http.StatusBadRequest)
		return
	}

	// Conditional AI Logic
	var subtasksParam pqtype.NullRawMessage
	if r.FormValue("use_ai") == "true" {
		subtasksJSON := logic.GenerateSubtasks(title, descRaw)
		subtasksParam = pqtype.NullRawMessage{RawMessage: subtasksJSON, Valid: true}
	} else {
		subtasksParam = pqtype.NullRawMessage{Valid: false}
	}

	_, err = s.Q.CreateTask(r.Context(), db.CreateTaskParams{
		Title:        title,
		Description:  descParam,
		OwnerID:      ownerIDParam,
		AssigneeText: assigneeParam,
		Subtasks:     subtasksParam,
		Priority:     int32(priorityInt),
		DueDate:      dateParam,
		Tags:         []string{},
		EventID:      eventUUID,
		Category:     category,
	})
	if err != nil {
		http.Error(w, "Error creating task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/events/"+eventIDStr, http.StatusSeeOther)
}

// 4) EDIT TASK (GET)
// 4) EDIT TASK (GET)
func (s *Server) handleEditTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	task, err := s.Q.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Task not found", 404)
		return
	}

	people, _ := s.Q.ListPeople(r.Context())

	// Parse subtasks
	var parsedSubtasks []logic.Subtask
	if task.Subtasks.Valid {
		_ = json.Unmarshal(task.Subtasks.RawMessage, &parsedSubtasks)
	}

	// --- NEW: Generate Rich Google Calendar Link ---
	calLink := ""
	if task.DueDate.Valid {
		day := task.DueDate.Time.Format("20060102")

		// Create a description that includes the main context AND the subtasks
		descText := fmt.Sprintf("CONTEXT:\n%s\n\nACTION PLAN:", task.Description.String)
		for _, sub := range parsedSubtasks {
			status := "[ ]"
			if sub.IsDone {
				status = "[x]"
			}
			descText += fmt.Sprintf("\n%s %s", status, sub.Title)
		}
		descText += "\n\n(Generated by SBF-OS)"

		// Build the URL
		calLink = fmt.Sprintf(
			"https://calendar.google.com/calendar/render?action=TEMPLATE&text=%s&details=%s&dates=%s/%s",
			url.QueryEscape("DEADLINE: "+task.Title), // Adds "DEADLINE:" to title for urgency
			url.QueryEscape(descText),
			day, day,
		)
	}
	// ------------------------------------------

	tmpl := s.templates["edit_task"]
	if tmpl == nil {
		http.Error(w, "Template not found", 500)
		return
	}

	data := struct {
		Task     db.Task
		People   []db.Person
		Subtasks []logic.Subtask
		GCalLink string
	}{
		Task:     task,
		People:   people,
		Subtasks: parsedSubtasks,
		GCalLink: calLink,
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

// 5) UPDATE TASK (POST)
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	ctx := r.Context()
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid task id", 400)
		return
	}

	// 1. Gather Basic Fields
	title := r.FormValue("title")
	desc := r.FormValue("description")
	status := r.FormValue("status")
	category := r.FormValue("category")
	priorityStr := r.FormValue("priority")
	dateStr := r.FormValue("due_date")
	ownerIDStr := r.FormValue("owner_id")
	assigneeName := r.FormValue("assignee_text")

	var titleParam sql.NullString
	if title != "" {
		titleParam = sql.NullString{String: title, Valid: true}
	}
	var descParam sql.NullString
	if desc != "" {
		descParam = sql.NullString{String: desc, Valid: true}
	}
	var statusParam sql.NullString
	if status != "" {
		statusParam = sql.NullString{String: status, Valid: true}
	}
	var categoryParam sql.NullString
	if category != "" {
		categoryParam = sql.NullString{String: category, Valid: true}
	}
	var assigneeParam sql.NullString
	if assigneeName != "" {
		assigneeParam = sql.NullString{String: assigneeName, Valid: true}
	}

	var priorityParam sql.NullInt32
	if priorityStr != "" {
		p, _ := strconv.Atoi(priorityStr)
		priorityParam = sql.NullInt32{Int32: int32(p), Valid: true}
	}
	var dateParam sql.NullTime
	if dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dateParam = sql.NullTime{Time: t, Valid: true}
		}
	}
	var ownerIDParam uuid.NullUUID
	if ownerIDStr != "" {
		if p, err := uuid.Parse(ownerIDStr); err == nil {
			ownerIDParam = uuid.NullUUID{UUID: p, Valid: true}
		}
	}

	// 2. Gather Existing/Manual Subtasks from Form
	subTitles := r.Form["subtask_title"]
	var currentSubtasks []logic.Subtask

	for i, t := range subTitles {
		if t == "" {
			continue
		}
		isDone := r.FormValue(fmt.Sprintf("subtask_done_%d", i)) == "on"
		currentSubtasks = append(currentSubtasks, logic.Subtask{
			Title:  t,
			IsDone: isDone,
		})
	}

	// 3. CHECK FOR AI TRIGGER (Updated with Debugging)
	if r.FormValue("action") == "generate_ai" {
		aiJSON := logic.GenerateSubtasks(title, desc)
		var aiSteps []logic.Subtask
		if err := json.Unmarshal(aiJSON, &aiSteps); err == nil {
			// Append AI steps to whatever is currently in the list
			currentSubtasks = append(currentSubtasks, aiSteps...)
		} else {
			// PRINT ERROR TO TERMINAL
			fmt.Println("❌ Handler JSON Error: Could not unmarshal AI response:", err)
			fmt.Println("   Raw JSON was:", string(aiJSON))
		}
	}
	// 4. Serialize Subtasks
	var subtasksParam pqtype.NullRawMessage
	if len(currentSubtasks) > 0 {
		b, _ := json.Marshal(currentSubtasks)
		subtasksParam = pqtype.NullRawMessage{RawMessage: b, Valid: true}
	} else {
		subtasksParam = pqtype.NullRawMessage{Valid: false}
	}

	// 5. Database Transaction
	txErr := s.Q.RunTx(ctx, s.DB, func(qtx *db.Queries) error {
		oldTask, err := qtx.GetTask(ctx, taskID)
		if err != nil {
			return err
		}

		newTask, err := qtx.UpdateTask(ctx, db.UpdateTaskParams{
			ID:           taskID,
			Title:        titleParam,
			Description:  descParam,
			Status:       statusParam,
			Priority:     priorityParam,
			DueDate:      dateParam,
			Category:     categoryParam,
			OwnerID:      ownerIDParam,
			AssigneeText: assigneeParam,
			Subtasks:     subtasksParam,
		})
		if err != nil {
			return err
		}

		diff := logic.CalculateChanges(oldTask, newTask)
		return qtx.CreateTaskEvent(ctx, db.CreateTaskEventParams{
			TaskID:    taskID,
			EventType: "UPDATED",
			Changes:   diff,
		})
	})

	if txErr != nil {
		http.Error(w, "Update failed: "+txErr.Error(), 500)
		return
	}

	// If we just generated AI, go back to Edit page to show them.
	// Otherwise, go to where they came from (Dashboard/Event View).
	if r.FormValue("action") == "generate_ai" {
		http.Redirect(w, r, fmt.Sprintf("/tasks/%s/edit", taskID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	}
}

// 6) VIEW HISTORY
func (s *Server) handleTaskEvents(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid task id", 400)
		return
	}
	events, err := s.Q.GetTaskEvents(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Failed to fetch events", 500)
		return
	}

	type EventView struct {
		EventType string
		CreatedAt string
		Changes   string
	}
	eventViews := make([]EventView, 0, len(events))
	for _, e := range events {
		eventViews = append(eventViews, EventView{
			EventType: e.EventType,
			CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
			Changes:   string(e.Changes),
		})
	}
	tmpl := s.templates["task_events"]
	if tmpl != nil {
		tmpl.ExecuteTemplate(w, "base", struct{ Events []EventView }{Events: eventViews})
	}
}

// 7) DELETE TASK
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}
	if err := s.Q.SoftDeleteTask(r.Context(), taskID); err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 8) CREATE EVENT
func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates, _ := s.Q.ListTemplates(r.Context())
		data := struct{ Templates []db.Template }{Templates: templates}
		tmpl := s.templates["create_event"]
		if tmpl != nil {
			tmpl.ExecuteTemplate(w, "base", data)
		}
		return
	}
	name := r.FormValue("name")
	dateStr := r.FormValue("event_date")
	templateIDStr := r.FormValue("template_id")
	eventDate, _ := time.Parse("2006-01-02", dateStr)

	event, err := s.Q.CreateEvent(r.Context(), db.CreateEventParams{
		Name:      name,
		EventDate: eventDate,
	})
	if err != nil {
		http.Error(w, "Failed to create event", 500)
		return
	}

	if templateIDStr != "" {
		tmplID, _ := uuid.Parse(templateIDStr)
		tmplTasks, _ := s.Q.GetTemplateTasks(r.Context(), tmplID)

		for _, t := range tmplTasks {
			dueDate := event.EventDate.AddDate(0, 0, -int(t.RelativeDueDays.Int32))
			s.Q.CreateTask(r.Context(), db.CreateTaskParams{
				Title:       t.Title,
				Priority:    t.Priority,
				Category:    t.Category,
				EventID:     event.ID,
				DueDate:     sql.NullTime{Time: dueDate, Valid: true},
				Description: t.Description,
			})
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 9) EDIT EVENT
func (s *Server) handleEditEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}
	event, err := s.Q.GetEvent(r.Context(), eventID)
	if err != nil {
		http.Error(w, "Event not found", 404)
		return
	}
	tmpl := s.templates["edit_event"]
	if tmpl != nil {
		tmpl.ExecuteTemplate(w, "base", struct{ Event db.Event }{Event: event})
	}
}

// 10) UPDATE EVENT
func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}
	name := r.FormValue("name")
	dateStr := r.FormValue("event_date")
	loc := r.FormValue("location")
	sum := r.FormValue("summary")

	var nameParam sql.NullString
	if name != "" {
		nameParam = sql.NullString{String: name, Valid: true}
	}
	var dateParam sql.NullTime
	if dateStr != "" {
		t, _ := time.Parse("2006-01-02", dateStr)
		dateParam = sql.NullTime{Time: t, Valid: true}
	}
	var locParam sql.NullString
	if loc != "" {
		locParam = sql.NullString{String: loc, Valid: true}
	}
	var sumParam sql.NullString
	if sum != "" {
		sumParam = sql.NullString{String: sum, Valid: true}
	}

	_, err = s.Q.UpdateEvent(r.Context(), db.UpdateEventParams{
		ID:        eventID,
		Name:      nameParam,
		EventDate: dateParam,
		Location:  locParam,
		Summary:   sumParam,
	})
	if err != nil {
		http.Error(w, "Update failed: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/events/"+eventID.String(), http.StatusSeeOther)
}

// 11) BATCH DELETE
func (s *Server) handleBatchDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	idStrings := r.Form["task_ids"]
	ids := make([]uuid.UUID, 0, len(idStrings))
	for _, idStr := range idStrings {
		if uid, err := uuid.Parse(idStr); err == nil {
			ids = append(ids, uid)
		}
	}
	if len(ids) > 0 {
		if err := s.Q.BatchSoftDeleteTasks(r.Context(), ids); err != nil {
			http.Error(w, "Batch delete failed: "+err.Error(), 500)
			return
		}
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
