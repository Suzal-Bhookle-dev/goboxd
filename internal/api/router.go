package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/config"
)

func RegisterRoutes(app *fiber.App, cfg *config.Config) {
	h := &Handler{Cfg: cfg}

	app.Get("/healthz", h.HandleHealthz)
	app.Post("/run", h.HandleRun)
}

type Handler struct {
	Cfg *config.Config
}
