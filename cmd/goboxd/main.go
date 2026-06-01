package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/api"
	"github.com/thesouldev/goboxd/internal/config"
)

func cleanStaleJails() {
	files, _ := os.ReadDir(os.TempDir())
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "gobox-") {
			os.RemoveAll(filepath.Join(os.TempDir(), f.Name()))
		}
	}
}

func main() {
	cleanStaleJails()

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
