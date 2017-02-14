package server

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"sync"
	"time"
)

type RoomClient interface {
	io.Writer

	// FLush flushes data to client side
	Flush()
}

type Room interface {
	// Id returns room's identity.
	Id() string

	// Serve starts handling room activites, returns on error.
	Serve() error

	// Stop stops handling room activites.
	Stop() error

	// ClaimSeats claims a seat from room, returns seat id.
	ClaimSeats() (int, error)

	// SetClient binds a RoomClient to seat.
	SetClientStream(seatId int, client RoomClient) error

	// RemoveClient removes a client from seat id.
	RemoveClient(seatId int) error

	// BroadcastStream streams to all room clients.
	BroadcastStream(fromSeatId int, stream []byte) error
}

const (
	roomStateStarted = "started"
	roomStateStopped = "stopped"
)

var (
	errRoomStarted = errors.New("room already started")
	errRoomIsFull  = errors.New("room is full")
)

type roomOptSetter func(r *room) error

func WithRoomRegistry(reg RoomRegistry) roomOptSetter {
	return func(r *room) error {
		if err := reg.AddRoom(r); err != nil {
			return err
		}
		r.registry = reg
		return nil
	}
}

type room struct {
	id                  string
	cap                 int
	nextAvailableSeatId int
	clients             map[int]RoomClient
	registry            RoomRegistry

	state string
	funcC chan func(*room)

	lock *sync.Mutex
}

const (
	roomCap = 8
	maxRoom = 341592653
)

func NewRoom(opts ...roomOptSetter) (*room, error) {
	r := &room{
		id:  fmt.Sprintf("r%d", randInt(maxRoom)),
		cap: roomCap,
		// seatId starts from 1
		nextAvailableSeatId: 1,
		clients:             make(map[int]RoomClient, roomCap),

		state: roomStateStopped,
		funcC: make(chan func(*room)),

		lock: &sync.Mutex{},
	}

	for seatId := 1; seatId <= r.cap; seatId++ {
		r.clients[seatId] = nil
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r room) Id() string { return r.id }

func (r *room) Serve() error {
	r.lock.Lock()

	if r.state != roomStateStopped {
		r.lock.Unlock()
		return errRoomStarted
	}

	r.state = roomStateStarted
	r.lock.Unlock()

	aliveTicker := time.NewTicker(180 * time.Second)
	defer aliveTicker.Stop()

	for r.state == roomStateStarted {
		select {
		case f := <-r.funcC:
			f(r)
		case <-aliveTicker.C:
			emptyRoom := true
			for seatId := 1; seatId < r.cap; seatId++ {
				if r.clients[seatId] != nil {
					emptyRoom = false
					break
				}
			}
			if emptyRoom {
				log.Printf("room %s became inactive, stopped", r)
				return r.stop()
			}
		}
	}
	return nil
}

func (r *room) stop() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.state != roomStateStarted {
		return nil
	}

	if r.registry != nil {
		r.registry.RemoveRoomById(r.id)
	}

	r.state = roomStateStopped
	return nil
}

func (r room) Stop() error {
	err := make(chan error)

	r.funcC <- func(r *room) {
		err <- r.stop()
	}

	return <-err
}

func (r *room) ClaimSeats() (int, error) {
	seatId := make(chan int)
	err := make(chan error)

	r.funcC <- func(r *room) {
		for s := r.nextAvailableSeatId; s <= r.cap; s++ {
			if r.clients[s] == nil {
				r.nextAvailableSeatId = s + 1
				seatId <- s
				err <- nil
				return
			}
		}

		seatId <- -1
		err <- errRoomIsFull
		return
	}

	return <-seatId, <-err
}

func (r *room) SetClientStream(seatId int, client RoomClient) error {
	err := make(chan error)

	r.funcC <- func(r *room) {
		r.clients[seatId] = client
		err <- nil
	}

	return <-err
}

func (r *room) RemoveClient(seatId int) error {
	err := make(chan error)

	r.funcC <- func(r *room) {
		r.clients[seatId] = nil
		if seatId < r.nextAvailableSeatId {
			r.nextAvailableSeatId = seatId
		}
		err <- nil
	}

	return <-err
}

func (r *room) BroadcastStream(fromSeatId int, stream []byte) error {
	err := make(chan error)

	r.funcC <- func(r *room) {
		if r.clients[fromSeatId] == nil {
			err <- fmt.Errorf("seat %d not in room", fromSeatId)
			return
		}

		var allWriteErr error

		for _, client := range r.clients {
			if client == nil {
				continue
			}
			if _, writeErr := client.Write(stream); writeErr != nil {
				// TODO combine errors
				log.Printf("BroadcastStream to %d %+v", fromSeatId, writeErr)
				allWriteErr = writeErr
			}
			client.Flush()
		}

		err <- allWriteErr
	}

	return <-err
}

func (r room) String() string { return r.id }

func (r room) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"id": r.id})
}

func randInt(max int64) int64 {
	b, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}

	return b.Int64()
}
