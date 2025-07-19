package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/tiredkangaroo/keylock/web/assets"
	"github.com/tiredkangaroo/keylock/web/views"
)

func SetGroup(router fiber.Router) {
	router.Use("/assets", filesystem.New(filesystem.Config{
		Root: http.FS(assets.Assets),
	}))
	router.Get("/", adaptor.HTTPHandler(templ.Handler(views.Main())))
	router.Get("/login", adaptor.HTTPHandler(templ.Handler(views.Login())))

	router.Get("/signup", adaptor.HTTPHandler(templ.Handler(views.Signup())))
}
