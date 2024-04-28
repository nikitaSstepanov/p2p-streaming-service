package dto

type ChunkDto struct {
	Buffer     []byte   `json:"buffer"`
	NextIndex   uint64  `json:"next"`
	FileVersion uint64  `json:"version"`
}