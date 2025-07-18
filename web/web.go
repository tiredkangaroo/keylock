package web

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/tiredkangaroo/keylock/web/views"
)

func SetGroup(router fiber.Router) {
	router.Get("/", adaptor.HTTPHandler(templ.Handler(views.Home("aji"))))
}
