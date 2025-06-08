package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func (app *application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(app.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		userId := claims["userId"].(float64)

		user, err := app.models.Users.Get(int(userId))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Set("user", user)

		c.Next()
	}
}

/**
Explanation of the Middleware

Retrieve the Authorization Header: The middleware starts by reading the Authorization header from the incoming request. This header should contain the JWT. If the header is missing, the middleware responds with a 401 Unauthorized status and stops further processing by calling c.Abort().
Extract the Bearer Token: The JWT is expected to be in the format Bearer {token}. The middleware removes the Bearer prefix to extract the actual token. If the token is not in the expected format, it responds with a 401 Unauthorized status and aborts the request.
Parse and Validate the JWT: The middleware uses the jwt.Parse function to decode and validate the token. It checks that the token’s signing method is HMAC, which is a common method for signing JWTs. The jwtSecret is used to verify the token’s signature, ensuring it hasn’t been tampered with.
Handle Invalid Tokens: If the token is invalid or an error occurs during parsing, the middleware responds with a 401 Unauthorized status and aborts the request.
Extract User Information: If the token is valid, the middleware extracts the user ID from the token’s claims and retrieves the corresponding user from the database. The user is then set in the request context using c.Set("user", user). This allows other handlers in the chain to access the authenticated user.
Allow the Request to Proceed: If the token is valid, the middleware calls c.Next(), allowing the request to proceed to the next handler in the chain.

*/
