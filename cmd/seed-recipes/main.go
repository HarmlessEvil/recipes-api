package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

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
		recipe.ID = primitive.NewObjectID()
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

	mongoDBClient, err := connectToMongoDB(ctx)
	if err != nil {
		return err
	}

	recipesCollection := mongoDBClient.Database(os.Getenv("MONGO_DATABASE")).Collection("stepByStepRecipes")

	return seedDatabase(ctx, recipesCollection)
}

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}
