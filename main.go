package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"golang.org/x/exp/slices"

	_ "embed"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

//go:embed recipes.json
var recipesJSON []byte

var recipes []Recipe

func init() {
	_ = json.Unmarshal(recipesJSON, &recipes)
}

func newRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	recipes = append(recipes, recipe)

	c.JSON(http.StatusOK, recipe)
}

func listRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func updateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	index := slices.IndexFunc(recipes, func(recipe Recipe) bool {
		return recipe.ID == id
	})
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})

		return
	}

	recipes[index] = recipe

	c.JSON(http.StatusOK, recipe)
}

func deleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	index := slices.IndexFunc(recipes, func(recipe Recipe) bool {
		return recipe.ID == id
	})
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})

		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func main() {
	router := gin.Default()
	router.POST("/recipes", newRecipeHandler)
	router.GET("/recipes", listRecipesHandler)
	router.PUT("/recipes/:id", updateRecipeHandler)
	router.DELETE("/recipes/:id", deleteRecipeHandler)

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
