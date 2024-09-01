package state

import (
	"sync"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
)

type State struct {
	movies map[uint64]*movie
	mutex  *sync.Mutex
}

type movie struct {
	torrent *decode.Torrent
	expires time.Time
}

func New() *State {
	movies := make(map[uint64]*movie)

	var mutex sync.Mutex

	go clearState(movies, &mutex)

	return &State{
		movies: movies,
		mutex:  &mutex,
	}
}

func (s *State) Get(id uint64) *decode.Torrent {
	s.mutex.Lock()
	
	movie, isFound := s.movies[id]

	s.mutex.Unlock()

	if !isFound {
		return nil
	}

	return movie.torrent
}

func (s *State) Add(id uint64, torrent *decode.Torrent, expires time.Duration) {
	new := &movie{
		torrent: torrent,
		expires: time.Now().Add(expires),
	}

	s.mutex.Lock()

	s.movies[id] = new

	s.mutex.Unlock()
}

func (s *State) ChangeExpires(id uint64, expires time.Duration) {
	s.mutex.Lock()

	movie, isFound := s.movies[id]
	
	if !isFound {
		s.mutex.Unlock()
		return
	}

	movie.expires = time.Now().Add(expires)

	s.mutex.Unlock()
}

func clearState(movies map[uint64]*movie, mutex *sync.Mutex) {
	for {
		mutex.Lock()

		for key := range movies {
			movie := movies[key]
			
			if movie.expires.Before(time.Now()) {
				delete(movies, key)
			}
		}

		mutex.Unlock()
		
		time.Sleep(10 * time.Minute)
	}
}