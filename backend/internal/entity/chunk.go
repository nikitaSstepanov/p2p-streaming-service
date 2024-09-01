package entity

type Chunk struct {
	Buffer     []byte 
	NextIndex   int
	FileVersion int 
}