package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/tiredkangaroo/keylock/database"
	"github.com/tiredkangaroo/keylock/web/assets"
	"github.com/tiredkangaroo/keylock/web/views"
)

func SetGroup(db *database.DB, sessionMiddleware fiber.Handler, router fiber.Router) {
	router.Use("/assets", filesystem.New(filesystem.Config{
		Root: http.FS(assets.Assets),
	}))
	router.Get("/", adaptor.HTTPHandler(templ.Handler(views.Main())))
	router.Get("/access", sessionMiddleware, adaptor.HTTPHandler(templ.Handler(views.Access())))
	router.Get("/login", adaptor.HTTPHandler(templ.Handler(views.Login())))
	router.Get("/signup", adaptor.HTTPHandler(templ.Handler(views.Signup())))
	router.Get("/home", sessionMiddleware, func(c *fiber.Ctx) error {
		user := c.Locals("user").(*database.User)
		pwds, err := db.ListPasswords(user.ID)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("error fetching passwords: " + err.Error())
		}
		fmt.Println("31", pwds)
		c.Set("Content-Type", fiber.MIMETextHTMLCharsetUTF8)
		return views.Home(user, pwds).Render(context.Background(), c.Response().BodyWriter())
	})
}
