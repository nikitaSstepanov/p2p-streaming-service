package dto


type ChunkDto struct {
	Buffer int    `json:"buffer"`
	NextIndex uint32 `json:"next"`
}