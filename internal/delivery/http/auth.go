package http

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"rkmin/internal/domain"
	"rkmin/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const userContextKey = "auth_user"

type JWTService struct {
	secret []byte
	repo   *repository.Repositories
}

func NewJWTService(secret string, repo *repository.Repositories) *JWTService {
	return &JWTService{secret: []byte(secret), repo: repo}
}

func (j *JWTService) Sign(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"id":       strconv.FormatUint(uint64(user.ID), 10),
		"email":    user.Email,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
}

func (j *JWTService) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("token")
		if tokenString == "" {
			auth := c.GetHeader("Authorization")
			tokenString = strings.TrimPrefix(auth, "Bearer ")
		}
		if tokenString == "" {
			fail(c, 401, c.Request.Method, errors.New("Unauthorized"))
			c.Abort()
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return j.secret, nil
		})
		if err != nil || !token.Valid {
			fail(c, 401, c.Request.Method, errors.New("Unauthorized"))
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fail(c, 401, c.Request.Method, errors.New("Unauthorized"))
			c.Abort()
			return
		}
		rawID, _ := claims["id"].(string)
		id64, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			fail(c, 401, c.Request.Method, errors.New("Unauthorized"))
			c.Abort()
			return
		}
		user, err := j.repo.FindUserByID(uint(id64))
		if err != nil {
			fail(c, 401, c.Request.Method, errors.New("Unauthorized"))
			c.Abort()
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func authUser(c *gin.Context) domain.User {
	user, _ := c.Get(userContextKey)
	return user.(domain.User)
}
