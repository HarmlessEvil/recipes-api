package handlers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	ctx        context.Context
	collection *mongo.Collection
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{ctx: ctx, collection: collection}
}

func (a *AuthHandler) AuthMiddleware() (gin.HandlerFunc, error) {
	issuerURL, err := url.Parse(fmt.Sprintf("https://%s/", os.Getenv("AUTH0_DOMAIN")))
	if err != nil {
		return nil, err
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
	)
	if err != nil {
		return nil, err
	}

	middleware := jwtmiddleware.New(jwtValidator.ValidateToken)
	return adapter.Wrap(middleware.CheckJWT), nil
}
