package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/api"
)

func main() {
	App := fiber.New()

	api.RegisterRoutes(App)

	log.Fatal(App.Listen(":8080"))
}
