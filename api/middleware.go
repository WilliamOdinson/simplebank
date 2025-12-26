package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/WilliamOdinson/simplebank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the authorization header
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("authorization header is not provided")))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("authorization header is not provided")))
			return
		}

		// Split the authorization header into fields
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("invalid authorization header format")))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("invalid authorization header format")))
			return
		}

		// Check the authorization type
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("unsupported authorization type: %s", authorizationType)))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("unsupported authorization type: %s", authorizationType)))
			return
		}

		// Verify the access token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// Set the authorization payload in the context for the next handlers
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
