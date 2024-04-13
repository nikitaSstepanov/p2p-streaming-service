package dto

type ChunkDto struct {
	Buffer     []byte  `json:"buffer"`
	NextIndex  uint64  `json:"next"`
}