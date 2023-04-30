package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/harmlessevil/recipes-api/handlers"
	"github.com/harmlessevil/recipes-api/models"
)

func setupRouter(t *testing.T) *gin.Engine {
	ctx := context.Background()

	mongoDBClient, err := connectToMongoDB(ctx)
	require.NoError(t, err)

	redisClient, err := connectToRedis(ctx)
	require.NoError(t, err)

	c := mongoDBClient.Database(os.Getenv("MONGO_DATABASE")).Collection("stepByStepRecipes")
	h := handlers.NewRecipesHandler(ctx, c, redisClient)

	router := gin.Default()

	router.GET("/recipes", h.ListRecipesHandler)
	router.POST("/recipes", h.NewRecipeHandler)
	router.PUT("/recipes/:id", h.UpdateRecipeHandler)
	router.GET("/recipes/:id", h.GetRecipeHandler)
	router.DELETE("/recipes/:id", h.DeleteRecipeHandler)

	return router
}

func connectToMongoDB(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

func connectToRedis(ctx context.Context) (*redis.Client, error) {
	redisOptions, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(redisOptions)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

func TestListRecipesHandler(t *testing.T) {
	ts := httptest.NewServer(setupRouter(t))
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/recipes", ts.URL))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var recipes []models.Recipe
	require.NoError(t, json.Unmarshal(data, &recipes))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 492, len(recipes))

	resp, err = http.Get(fmt.Sprintf("%s/recipes", ts.URL))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	data, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(data, &recipes))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 492, len(recipes))
}

func TestNewRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(setupRouter(t))
	defer ts.Close()

	data, err := json.Marshal(models.Recipe{
		Name: "New York Pizza",
	})
	require.NoError(t, err)

	resp, err := http.Post(fmt.Sprintf("%s/recipes", ts.URL), "application/json", bytes.NewReader(data))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(fmt.Sprintf("%s/recipes", ts.URL), "application/json", nil)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(setupRouter(t))
	defer ts.Close()

	data, err := json.Marshal(models.Recipe{
		Name: "Oregano Marinated Chicken",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/recipes/644bf0e2d9d9e29d5c6efad8", ts.URL), bytes.NewReader(data))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/recipes/644bf0e2d9d9e29d5c6efad8", ts.URL), nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/recipes/1", ts.URL), bytes.NewReader(data))
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/recipes/644bd54a533f211534d730b8", ts.URL), bytes.NewReader(data))
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(setupRouter(t))
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/recipes/644bf0e2d9d9e29d5c6efad8", ts.URL))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var actual models.Recipe
	require.NoError(t, json.Unmarshal(data, &actual))

	require.Equal(t, "Oregano Marinated Chicken", actual.Name)

	resp, err = http.Get(fmt.Sprintf("%s/recipes/1", ts.URL))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("%s/recipes/644bd54a533f211534d730b8", ts.URL))
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(setupRouter(t))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/recipes/644bf0e2d9d9e29d5c6efad8", ts.URL), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/recipes/1", ts.URL), nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/recipes/644bd54a533f211534d730b8", ts.URL), nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
