package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tiredkangaroo/keylock/cache"
	"github.com/tiredkangaroo/keylock/database"
)

var sessionExpiration time.Duration = time.Hour * 24 * 7

func newSessionForUser(userID int64) (string, error) {
	sessionIDRaw := make([]byte, 20)
	rand.Read(sessionIDRaw)
	sessionID := hex.EncodeToString(sessionIDRaw)

	err := cache.HSetWithExpiration("user-session", sessionID, strconv.Itoa(int(userID)), sessionExpiration)
	if err != nil {
		return "", fmt.Errorf("redis save error: %w", err)
	}

	return sessionID, nil
}

func getUser(c *fiber.Ctx) *database.User {
	return c.Locals("user").(*database.User)
}
