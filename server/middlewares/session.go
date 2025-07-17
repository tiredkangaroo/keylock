package middlewares

import (
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/cache"
	"github.com/tiredkangaroo/keylock/database"
)

func SessionMiddleware(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// NOTE: we should look over login required stuff
		// we have two ways for session tokens: cookie "session_token" or Authorization header
		// the cookie takes precedence over the header
		session_token := c.Cookies("session_token", c.Get("Authorization"))
		if session_token == "" {
			slog.Debug("no session token found in cookie or Authorization header")
			return c.Next()
		}
		userid_raw, err := cache.HGet("user-session", session_token)
		if err != nil {
			slog.Error("get session token from redis", "err", err)
			return c.Next()
		}
		userid, err := strconv.ParseInt(userid_raw, 10, 64)
		if err != nil {
			slog.Error("invalid session token, userid stored in redis is not an integer", "err", err) // shouldn't happen but i handle my erros :)
			return c.Next()
		}
		user, err := db.GetUserByID(userid)
		if err != nil {
			slog.Error("get user by id from database", "userid", userid, "err", err)
			return c.Next()
		}
		c.Locals("user", user)

		return c.Next()
	}
}
