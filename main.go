// Recipes API
//
// This is a sample recipes API.
// You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Alexander Chori <alexandrchori@gmail.com> http://chorilabs.com
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
	// swagger:operation GET /recipes recipes listRecipes
	//
	// Returns list of recipes
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//  '200':
	//   description: Successful operation

	c.JSON(http.StatusOK, recipes)
}

func updateRecipeHandler(c *gin.Context) {
	// swagger:operation PUT /recipes/{id} recipes updateRecipe
	//
	// Update an existing recipe
	//
	// ---
	// parameters:
	//   - name: id
	//     in: path
	//     description: ID of the recipe
	//     required: true
	//     type: string
	// produces:
	//   - application/json
	// responses:
	//  '200':
	//   description: Successful operation
	//  '400':
	//   description: Invalid input
	//  '404':
	//   description: Invalid recipe ID

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
	// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
	//
	// Delete an existing recipe
	//
	// ---
	// parameters:
	//   - name: id
	//     in: path
	//     description: ID of the recipe
	//     required: true
	//     type: string
	// produces:
	//   - application/json
	// responses:
	//  '200':
	//   description: Successful operation
	//  '400':
	//   description: Invalid input
	//  '404':
	//   description: Invalid recipe ID

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

func searchRecipesHandler(c *gin.Context) {
	// swagger:operation GET /recipes/search recipes searchRecipe
	//
	// Search for existing recipe by tag
	//
	// ---
	// parameters:
	//   - name: tag
	//     in: query
	//     description: tag of recipes
	//     required: true
	//     type: string
	// produces:
	//   - application/json
	// responses:
	//  '200':
	//   description: Successful operation
	//  '400':
	//   description: Invalid input
	//  '404':
	//   description: Invalid recipe ID

	tag := c.Query("tag")

	var res []Recipe
	for _, recipe := range recipes {
		if slices.ContainsFunc(recipe.Tags, func(t string) bool {
			return strings.EqualFold(t, tag)
		}) {
			res = append(res, recipe)
		}
	}

	c.JSON(http.StatusOK, res)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", newRecipeHandler)
	router.GET("/recipes", listRecipesHandler)
	router.PUT("/recipes/:id", updateRecipeHandler)
	router.DELETE("/recipes/:id", deleteRecipeHandler)
	router.GET("/recipes/search", searchRecipesHandler)

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
