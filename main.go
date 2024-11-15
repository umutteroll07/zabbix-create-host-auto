package main

import (
	"zabbix-create-host-auto/app/route"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	route.SetupRoutes(app)
	app.Listen(":3000")
}
