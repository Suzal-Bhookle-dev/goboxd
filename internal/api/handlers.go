package api

import (
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

	lang, _ := h.Cfg.GetLanguage(req.Language)

	resp, err := sandbox.Execute(*req, lang)
	if err != nil {
		return c.Status(500).JSON(errRes("internal_error", "Sandbox execution failed: "+err.Error()))
	}

	return c.Status(200).JSON(resp)
}
