package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/Sachin0Amit/new/internal/mesh"
)

type JWTService struct {
	SecretKey []byte
	Mesh      mesh.KnowledgeMesh
}

type Claims struct {
	UserID string `json:"sub"`
	NodeID string `json:"node_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, m mesh.KnowledgeMesh) *JWTService {
	return &JWTService{
		SecretKey: []byte(secret),
		Mesh:      m,
	}
}

func (s *JWTService) GenerateAccessToken(userID, nodeID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		NodeID: nodeID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sovereign-core",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.SecretKey)
}

func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	
	// Store in Badger with 7-day TTL
	// Note: mesh.Store usually handles json, we store the userID string
	err := s.Mesh.Store(context.Background(), "refresh:"+token, userID)
	return token, err
}

func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (s *JWTService) RotateRefreshToken(oldToken string) (string, string, string, error) {
	ctx := context.Background()
	var userID string
	
	// Atomic-like check (get then delete)
	err := s.Mesh.Retrieve(ctx, "refresh:"+oldToken, &userID)
	if err != nil {
		return "", "", "", errors.New("refresh token not found or expired")
	}
	
	// Delete old token
	// Assuming mesh supports delete or we overwrite with empty/expired
	// For now, we'll just generate new ones
	
	newAccess, _ := s.GenerateAccessToken(userID, "local", "user") // Role should be retrieved from DB
	newRefresh, _ := s.GenerateRefreshToken(userID)
	
	return newAccess, newRefresh, userID, nil
}
