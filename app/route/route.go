package route

import (
	"zabbix-create-host-auto/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Post("/create-hosts", controllers.CreateAutoHost)
}
