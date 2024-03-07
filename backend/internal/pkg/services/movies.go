package services

import (
	"github.com/gin-gonic/gin"
)

type Movies struct {}

func NewMovies() *Movies {
	return &Movies{}
}

func (m *Movies) GetMovies(ctx *gin.Context) {

}

func (m *Movies) GetMovieById(ctx *gin.Context) {

}

func (m *Movies) GetMovieChunck() {

}