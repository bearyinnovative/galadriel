package main

import (
	"net/http"

	"github.com/bearyinnovative/galadriel/server"
	"github.com/bearyinnovative/galadriel/server/api"
)

func main() {
	reg := server.NewRoomRegistry()
	route, _ := api.NewRoute(&api.RouteOptions{
		RoomRegistry: reg,
	})

	http.ListenAndServe(":8181", route)
}
