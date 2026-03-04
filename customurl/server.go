package customurl

import (
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	App *fiber.App
}

func New() *Server {
	app := fiber.New()

	return &Server{
		App: app,
	}
}

func (s *Server) Start(port string) error {
	return s.App.Listen(port)
}
