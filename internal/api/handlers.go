package api

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/models"
	"github.com/thesouldev/goboxd/internal/sandbox"
)

func (h *Handler) HandleHealthz(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) HandleRun(c fiber.Ctx) error {
	req := new(models.RunRequest)

	if err := c.Bind().JSON(req); err != nil {
		return c.Status(400).JSON(errRes("bad_json", "The request body is not valid JSON"))
	}

	if errResp := h.ValidateRequest(req); errResp != nil {
		return c.Status(400).JSON(errResp)
	}

	lang, ok := h.Cfg.GetLanguage(req.Language)
	if !ok {
		return c.Status(400).JSON(errRes("unsupported_language", "Language is not supported"))
	}
	
	buildCmd := "none"
	if lang.Build != nil {
		buildCmd = lang.Build.Cmd
	}
	log.Printf("DEBUG: Running language %s with build command: %s", lang.ID, buildCmd)

	resp, err := sandbox.Execute(*req, lang)
	if err != nil {
		return c.Status(500).JSON(errRes("internal_error", "Sandbox execution failed: "+err.Error()))
	}

	return c.Status(200).JSON(resp)
}
