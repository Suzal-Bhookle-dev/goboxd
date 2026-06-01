package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/api"
	"github.com/thesouldev/goboxd/internal/config"
)

func main() {
	App := fiber.New()

	cfg, err := config.Load("languages.yaml")
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	api.RegisterRoutes(App, cfg)

	log.Println("Starting server on port 8080...")

	err = App.Listen(":8080")
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
