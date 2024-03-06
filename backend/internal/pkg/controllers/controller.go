package controllers

import (

	"github.com/gin-gonic/gin"

)

type Controller struct {
	
}

func New() *Controller {

	return &Controller {}
	
}

func (c *Controller) InitRoutes() *gin.Engine {

	router := gin.New()

	//TODO: add routes.

	return router

}