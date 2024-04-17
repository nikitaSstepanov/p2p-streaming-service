package state

import (
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/bittorrent/decode"
)

type State struct {
	Movies *map[uint64]Movie
}

type Movie struct {
	Torrent *decode.Torrent
	Expires time.Time
}

func New() *State {
	movies := make(map[uint64]Movie)

	go clearState(&movies)

	return &State{
		Movies: &movies,
	}
}

func clearState(movies *map[uint64]Movie) error {
	for {
		time.Sleep(10 * time.Minute)
	}
}

