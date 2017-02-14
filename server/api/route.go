package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bearyinnovative/galadriel/server"
)

type RouteOptions struct {
	RoomRegistry server.RoomRegistry
	APIPrefix    string
}

func NewRoute(opt *RouteOptions) (http.Handler, error) {
	if opt.RoomRegistry == nil {
		return nil, errors.New("RoomRegistry is required")
	}
	reg := opt.RoomRegistry

	if opt.APIPrefix == "" {
		opt.APIPrefix = "/api/v1"
	}

	api := func(r string) string {
		return fmt.Sprintf("%s%s", opt.APIPrefix, r)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		api("/room.create"),
		func(w http.ResponseWriter, r *http.Request) {
			CreateRoom(w, r, reg)
		},
	)
	mux.HandleFunc(
		api("/room.stream"),
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				GetStream(w, r, reg)
			} else {
				StreamRoom(w, r, reg)
			}
		},
	)
	mux.HandleFunc(
		api("/room.join"),
		func(w http.ResponseWriter, r *http.Request) {
			JoinRoom(w, r, reg)
		},
	)

	return mux, nil
}
