package api

import (
	"github.com/gofiber/fiber/v3"
)

func HandleHealthz(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
