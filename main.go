// Recipes API
//
// This is a sample recipes API.
// You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin
//
// Schemes: http
// Host: api.recipes.io:8080
// BasePath: /
// Version: 1.0.0
// Contact: Alexander Chori <alexandrchori@gmail.com> http://chorilabs.com
// SecurityDefinitions:
// api_key:
//
//	type: apiKey
//	name: Authorization
//	in: header
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
	"encoding/json"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/harmlessevil/recipes-api/handlers"
	"github.com/harmlessevil/recipes-api/models"

	_ "embed"
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

func connectToRedis(ctx context.Context) (*redis.Client, error) {
	redisOptions, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(redisOptions)

	status := redisClient.Ping(ctx)
	log.Println("Redis", status)

	return redisClient, nil
}

func runMain() error {
	ctx := context.Background()

	mongoDBClient, err := connectToMongoDB(ctx)
	if err != nil {
		return err
	}

	redisClient, err := connectToRedis(ctx)
	if err != nil {
		return err
	}

	recipesCollection := mongoDBClient.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	usersCollection := mongoDBClient.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	authHandler := handlers.NewAuthHandler(ctx, usersCollection)
	recipesHandler := handlers.NewRecipesHandler(ctx, recipesCollection, redisClient)

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}

	router.Use(cors.New(corsConfig))

	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.GET("/recipes/:id", recipesHandler.GetRecipeHandler)
	router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	authenticated := router.Group("/")

	authMiddleware, err := authHandler.AuthMiddleware()
	if err != nil {
		return err
	}

	authenticated.Use(authMiddleware)
	{
		authenticated.POST("/recipes", recipesHandler.NewRecipeHandler)
		authenticated.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authenticated.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	}

	return router.Run()
}

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}
