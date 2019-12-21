package server

func (s *Server) setupRoutes() {
	s.router.Get("/friend/{id}", s.HandleGetFriend)
	s.router.Post("/friend", s.HandleCreateFriend)
	s.router.Delete("/friend/{id}", s.HandleDeleteFriend)
	s.router.Patch("/friend", s.HandleUpdateFriend)
}
