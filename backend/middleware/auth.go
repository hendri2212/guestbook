package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const authContextKey contextKey = "auth"

type AuthClaims struct {
	AdminUserID string `json:"admin_user_id"`
	CompanyID   string `json:"company_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	secret []byte
}

type authErrorResponse struct {
	Error string `json:"error"`
}

func NewAuthenticator(secret string) Authenticator {
	return Authenticator{secret: []byte(secret)}
}

func (auth Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ParseRequest(r)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (auth Authenticator) GenerateToken(adminUserID string, companyID string, email string, role string, expiresAt time.Time) (string, error) {
	claims := AuthClaims{
		AdminUserID: adminUserID,
		CompanyID:   companyID,
		Email:       email,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(auth.secret)
}

func (auth Authenticator) ParseRequest(r *http.Request) (*AuthClaims, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errors.New("authorization header is required")
	}

	tokenString := strings.TrimPrefix(header, "Bearer ")
	if tokenString == header || strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("bearer token is required")
	}

	claims := AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return auth.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return &claims, nil
}

func FromContext(ctx context.Context) (*AuthClaims, bool) {
	claims, ok := ctx.Value(authContextKey).(*AuthClaims)
	return claims, ok
}

func writeAuthError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(authErrorResponse{Error: message}); err != nil {
		log.Printf("failed to write auth error response: %v", err)
	}
}
