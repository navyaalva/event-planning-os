package server

func (s *Server) routes() {
	// 1. Dashboard
	s.Router.Get("/", s.handleDashboard)

	// 2. Event Management
	s.Router.Get("/events/new", s.handleCreateEvent)
	s.Router.Post("/events/new", s.handleCreateEvent)
	s.Router.Get("/events/{id}", s.handleEventDetail)
	s.Router.Get("/events/{id}/edit", s.handleEditEvent)
	s.Router.Post("/events/{id}/update", s.handleUpdateEvent)

	// 3. Task Creation
	s.Router.Get("/tasks/new", s.handleCreateTask)
	s.Router.Post("/tasks/new", s.handleCreateTask)

	// 4. Task Editing & Updates
	s.Router.Get("/tasks/{id}/edit", s.handleEditTask)
	s.Router.Post("/tasks/{id}/update", s.handleUpdateTask)
	s.Router.Post("/tasks/{id}/delete", s.handleDeleteTask)

	// 5. Batch Operations
	s.Router.Post("/tasks/batch-delete", s.handleBatchDelete)

	// 6. History
	s.Router.Get("/tasks/{id}/events", s.handleTaskEvents)

	// Legacy redirect
	s.Router.Get("/tasks", s.handleDashboard)
}
