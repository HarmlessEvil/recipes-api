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
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	_ "embed"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	// swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

//go:embed recipes.json
var recipesJSON []byte

var collection *mongo.Collection

func newRecipeHandler(c *gin.Context) {
	// swagger:operation POST /recipes recipes newRecipe
	//
	// Create new recipe
	//
	// ---
	// produces:
	//   - application/json
	// responses:
	//  '200':
	//   description: Successful operation
	//  '400':
	//   description: Invalid input

	ctx := context.TODO()

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	if _, err := collection.InsertOne(ctx, recipe); err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting recipe",
		})

		return
	}

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

	ctx := context.TODO()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		_ = cur.Close(ctx)
	}(cur, ctx)

	var recipes []Recipe
	for cur.Next(ctx) {
		var recipe Recipe
		if err := cur.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

func getRecipeHandler(c *gin.Context) {
	// swagger:operation GET /recipes/{id} recipes getRecipe
	//
	// Get an existing recipe
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
	//  '404':
	//   description: Invalid recipe ID

	ctx := context.TODO()

	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	var recipe Recipe
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&recipe); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Recipe not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, recipe)
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

	ctx := context.TODO()

	id := c.Param("id")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if _, err := collection.UpdateOne(ctx, bson.M{
		"_id": objectID,
	}, bson.D{{
		Key: "$set", Value: bson.D{
			{Key: "name", Value: recipe.Name},
			{Key: "instructions", Value: recipe.Instructions},
			{Key: "ingredients", Value: recipe.Ingredients},
			{Key: "tags", Value: recipe.Tags},
		},
	}}); err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been updated",
	})
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
	//  '404':
	//   description: Invalid recipe ID

	ctx := context.TODO()

	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	res, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})

		return
	}

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

	ctx := context.TODO()

	tag := c.Query("tag")

	opts := options.Find().SetCollation(&options.Collation{
		Locale:        "en_US",
		CaseLevel:     false,
		Normalization: true,
	})

	cur, err := collection.Find(ctx, bson.M{
		"tags": tag,
	}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		_ = cur.Close(ctx)
	}(cur, ctx)

	var recipes []Recipe
	for cur.Next(ctx) {
		var recipe Recipe
		if err := cur.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

func connectToDatabase(ctx context.Context) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	log.Println("Connected to MongoDB")

	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	return nil
}

func seedDatabase(ctx context.Context) error {
	var recipes []Recipe
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

	if err := connectToDatabase(ctx); err != nil {
		return err
	}

	router := gin.Default()
	router.POST("/recipes", newRecipeHandler)
	router.GET("/recipes", listRecipesHandler)
	router.GET("/recipes/:id", getRecipeHandler)
	router.PUT("/recipes/:id", updateRecipeHandler)
	router.DELETE("/recipes/:id", deleteRecipeHandler)
	router.GET("/recipes/search", searchRecipesHandler)

	return router.Run()
}

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}
