package main

import (
	"log/slog"

	"github.com/tiredkangaroo/keylock/database"
	"github.com/tiredkangaroo/keylock/server"
)

type Key interface {
	Value() ([]byte, error)
}

func main() {
	db, err := database.Database()
	if err != nil {
		slog.Error("connecting to database failed (fatal)", "error", err)
		return
	}
	defer db.Close()
	slog.Info("opened database")

	// user := &database.User{
	// 	Name: "testuser1",
	// }
	// err = db.SaveUser(user)
	// if err != nil {
	// 	slog.Error("saving user failed", "error", err)
	// 	return
	// }

	// slog.Info("saved user", "name", "testuser1")
	// err = db.SavePassword("172240", &database.Password{
	// 	UserID: user.ID,
	// 	Name:   "testpassword",
	// 	Value:  []byte("testpasswordvalue"),
	// })
	// if err != nil {
	// 	slog.Error("saving password failed", "error", err)
	// 	return
	// }
	// slog.Info("saved password", "name", "172240", "user_id", user.ID)

	// pwd := &database.Password{
	// 	UserID: user.ID,
	// 	Name:   "testpassword",
	// }
	// err = db.RetrievePassword("172240", pwd)
	// if err != nil {
	// 	slog.Error("retrieving password failed", "error", err)
	// 	return
	// }
	// slog.Info("retrieved password", "name", "172240", "user_id", user.ID, "value", string(pwd.Value))

	// err = fasthttp.Serve(listener, func(ctx *fasthttp.RequestCtx) {
	// })
	// if err != nil {
	// 	slog.Error("serving requests failed (fatal)", "addr", listener.Addr().String(), "error", err)
	// 	return
	// }
	s := &server.Server{}
	s.Init(db)
	if err := s.Start(); err != nil {
		slog.Error("server failed (fatal)", "error", err)
		return
	}
}
