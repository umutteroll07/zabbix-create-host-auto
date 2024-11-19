package main

import (
	"log"
	"zabbix-create-host-auto/app/route"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	app := fiber.New()
	route.SetupRoutes(app)
	log.Fatal(app.Listen(":3000"))
}
