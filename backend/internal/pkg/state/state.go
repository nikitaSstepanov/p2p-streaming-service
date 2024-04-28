package state

import (
	"sync"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
)

type State struct {
	Movies *map[uint64]Movie
	Mutex  *sync.Mutex
}

type Movie struct {
	Torrent *decode.Torrent
	Expires time.Time
}

func New() *State {
	movies := make(map[uint64]Movie)

	var mutex sync.Mutex

	go clearState(&movies, &mutex)

	return &State{
		Movies: &movies,
		Mutex: &mutex,
	}
}

func clearState(movies *map[uint64]Movie, mutex *sync.Mutex) error {
	for {
		mutex.Lock()

		for key := range *movies {
			movie := (*movies)[key]
			
			if movie.Expires.Before(time.Now()) {
				delete(*movies, key)
			}
		}

		mutex.Unlock()
		time.Sleep(10 * time.Minute)
	}
}

