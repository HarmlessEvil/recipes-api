package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	router := gin.Default()

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
