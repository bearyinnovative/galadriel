package server

import (
	"log"
	"sync"
)

type RoomRegistry interface {
	// GetRoomById retrieves room by its id, returns nil when not found
	GetRoomById(id string) Room

	// AddRoom adds room to registry
	AddRoom(room Room) error

	// RemoveRoomById removes room by its id
	RemoveRoomById(id string)
}

type roomRegistry struct {
	rooms map[string]Room

	lock *sync.RWMutex
}

func NewRoomRegistry() *roomRegistry {
	return &roomRegistry{
		rooms: make(map[string]Room),
		lock:  &sync.RWMutex{},
	}
}

func (r *roomRegistry) AddRoom(room Room) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.rooms[room.Id()] = room
	log.Printf("%s added to registry", room)
	return nil
}

func (r *roomRegistry) GetRoomById(id string) Room {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if room, exists := r.rooms[id]; exists {
		return room
	} else {
		return nil
	}
}

func (r *roomRegistry) RemoveRoomById(id string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.rooms, id)
}
