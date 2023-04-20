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
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/harmlessevil/recipes-api/handlers"
	"github.com/harmlessevil/recipes-api/models"
)

//go:embed recipes.json
var recipesJSON []byte

func connectToMongoDB(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")

	return client, nil
}

func seedDatabase(ctx context.Context, collection *mongo.Collection) error {
	var recipes []models.Recipe
	if err := json.Unmarshal(recipesJSON, &recipes); err != nil {
		return err
	}

	data := make([]any, len(recipes))
	for i, recipe := range recipes {
		data[i] = recipe
	}

	res, err := collection.InsertMany(ctx, data)
	if err != nil {
		return err
	}

	log.Println("Inserted recipes: ", len(res.InsertedIDs))

	return nil
}

func runMain() error {
	ctx := context.Background()

	client, err := connectToMongoDB(ctx)
	if err != nil {
		return err
	}

	recipesCollection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipesHandler := handlers.NewRecipesHandler(ctx, recipesCollection)

	router := gin.Default()
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.GET("/recipes/:id", recipesHandler.GetRecipeHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	return router.Run()
}

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}
