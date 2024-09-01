package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

type ChunkDto struct {
	Buffer     []byte   `json:"buffer"`
	NextIndex   int  `json:"next"`
	FileVersion int  `json:"version"`
}

type MovieDto struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func MovieToDto(movie *entity.Movie) MovieDto {
	return MovieDto{
		Id: movie.Id,
		Name: movie.Name,
	}
}

func ChunkToDto(chunk *entity.Chunk) ChunkDto {
	return ChunkDto{
		Buffer: chunk.Buffer,
		NextIndex: chunk.NextIndex,
		FileVersion: chunk.FileVersion,
	}
}