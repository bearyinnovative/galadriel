package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bearyinnovative/galadriel/server"
)

type HTTPRoomClient struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

func (c HTTPRoomClient) Write(data []byte) (int, error) {
	return c.writer.Write(data)
}

func (c HTTPRoomClient) Flush() {
	c.flusher.Flush()
}

func CreateRoom(w http.ResponseWriter, r *http.Request, reg server.RoomRegistry) {
	req := Req{w, r}
	if !req.requireMethod("POST") {
		return
	}

	room, err := server.NewRoom(server.WithRoomRegistry(reg))
	if err != nil {
		req.responseErr(err)
		return
	}

	go room.Serve()

	req.created(room)
	return
}

func StreamRoom(w http.ResponseWriter, r *http.Request, reg server.RoomRegistry) {
	req := Req{w, r}
	if !req.requireMethod("POST") {
		return
	}

	roomId, exists := req.queryValue("room_id")
	if !exists {
		req.badRequest(errors.New("room_id required"))
		return
	}
	room := reg.GetRoomById(roomId)
	if room == nil {
		req.notFound(fmt.Errorf("room %s not found", roomId))
		return
	}

	seatId, exists := req.queryValueInt("seat_id")
	if !exists {
		req.badRequest(errors.New("seat_id required"))
		return
	}

	stream, err := ioutil.ReadAll(r.Body)
	if err != nil {
		req.badRequest(err)
		return
	}
	r.Body.Close()
	room.BroadcastStream(seatId, stream)

	w.WriteHeader(204)
}

func JoinRoom(w http.ResponseWriter, r *http.Request, reg server.RoomRegistry) {
	req := Req{w, r}
	if !req.requireMethod("POST") {
		return
	}

	roomId, exists := req.queryValue("room_id")
	if !exists {
		req.badRequest(errors.New("room_id required"))
		return
	}
	room := reg.GetRoomById(roomId)
	if room == nil {
		req.notFound(fmt.Errorf("room %s not found", roomId))
		return
	}

	seatId, err := room.ClaimSeats()
	if err != nil {
		req.gone(err)
		return
	}

	req.created(mapResp{"seat_id": seatId})
	return
}

func GetStream(w http.ResponseWriter, r *http.Request, reg server.RoomRegistry) {
	req := Req{w, r}
	if !req.requireMethod("GET") {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		req.serverError(errors.New("unsupported client"))
		return
	}

	roomId, exists := req.queryValue("room_id")
	if !exists {
		req.badRequest(errors.New("room_id required"))
		return
	}
	room := reg.GetRoomById(roomId)
	if room == nil {
		req.notFound(fmt.Errorf("room %s not found", roomId))
		return
	}

	seatId, exists := req.queryValueInt("seat_id")
	if !exists {
		req.badRequest(errors.New("seat_id required"))
		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "video/webm")
	w.Header().Add("Cache-Control", "private")

	if err := room.SetClientStream(seatId, HTTPRoomClient{w, flusher}); err != nil {
		req.serverError(err)
		return
	}
}
