package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/utils"
)

type Claims struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

func (app *application) JWTAuthenticated(roles []models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, err := getTokenFromHeader(c)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("server.secret")), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			app.logger.Debug(err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)

		// If no roles are specified, accept all valid tokens
		if len(roles) == 0 {
			c.Next()
			return
		}

		validRoles := utils.IntersectString(models.RolesToStrings(roles), claims.Roles)

		if len(validRoles) == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()

	}
}

const HEADER_PREFIX string = "Bearer "

func getTokenFromHeader(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", models.ErrNoCredentails
	}
	if !strings.HasPrefix(header, HEADER_PREFIX) {
		return "", models.ErrNoCredentails
	}

	header = strings.TrimPrefix(header, HEADER_PREFIX)

	return header, nil
}
