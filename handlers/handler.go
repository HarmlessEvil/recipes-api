package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/harmlessevil/recipes-api/models"
)

type RecipesHandler struct {
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{ctx: ctx, collection: collection, redisClient: redisClient}
}

func (h *RecipesHandler) NewRecipeHandler(c *gin.Context) {
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

	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	if _, err := h.collection.InsertOne(h.ctx, recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting recipe",
		})

		return
	}

	log.Println("Remove data from Redis")
	if err := h.redisClient.Del(h.ctx, "recipes").Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting recipe",
		})

		return
	}

	c.JSON(http.StatusOK, recipe)
}

func (h *RecipesHandler) ListRecipesHandler(c *gin.Context) {
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

	val, err := h.redisClient.Get(h.ctx, "recipes").Result()
	if errors.Is(err, redis.Nil) {
		//log.Println("Request to MongoDB")

		cur, err := h.collection.Find(h.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
		defer func(cur *mongo.Cursor, ctx context.Context) {
			_ = cur.Close(ctx)
		}(cur, h.ctx)

		var recipes []models.Recipe
		for cur.Next(h.ctx) {
			var recipe models.Recipe
			if err := cur.Decode(&recipe); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}

			recipes = append(recipes, recipe)
		}

		data, err := json.Marshal(recipes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if err := h.redisClient.Set(h.ctx, "recipes", string(data), 0).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, recipes)
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	//log.Println("Request to Redis")

	var recipes []models.Recipe
	if err := json.Unmarshal([]byte(val), &recipes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, recipes)
}

func (h *RecipesHandler) GetRecipeHandler(c *gin.Context) {
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

	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	var recipe models.Recipe
	if err := h.collection.FindOne(h.ctx, bson.M{"_id": objectID}).Decode(&recipe); err != nil {
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

func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
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

	var recipe models.Recipe
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

	res, err := h.collection.UpdateOne(h.ctx, bson.M{
		"_id": objectID,
	}, bson.D{{
		Key: "$set", Value: bson.D{
			{Key: "name", Value: recipe.Name},
			{Key: "instructions", Value: recipe.Instructions},
			{Key: "ingredients", Value: recipe.Ingredients},
			{Key: "tags", Value: recipe.Tags},
		},
	}})

	if err != nil {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Recipe not found",
		})

		return
	}

	log.Println("Remove data from Redis")
	if err := h.redisClient.Del(h.ctx, "recipes").Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting recipe",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been updated",
	})
}

func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
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

	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	res, err := h.collection.DeleteOne(h.ctx, bson.M{"_id": objectID})
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

func (h *RecipesHandler) SearchRecipesHandler(c *gin.Context) {
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

	tag := c.Query("tag")

	opts := options.Find().SetCollation(&options.Collation{
		Locale:        "en_US",
		CaseLevel:     false,
		Normalization: true,
	})

	cur, err := h.collection.Find(h.ctx, bson.M{
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
	}(cur, h.ctx)

	var recipes []models.Recipe
	for cur.Next(h.ctx) {
		var recipe models.Recipe
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
