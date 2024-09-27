package security

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT operations including token generation and validation.
type JWTManager struct {
	AccessExp  time.Duration
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func NewJWTManager(accessDuration time.Duration, privateKeyPEM, publicKeyPEM []byte) (*JWTManager, error) {
	if accessDuration <= 0 {
		return nil, errors.New("durations must be positive")
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %w", err)
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %w", err)
	}

	return &JWTManager{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		AccessExp:  accessDuration,
	}, nil
}

// JWTClaims extends standard JWT claims with a UserID for user identification.
type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a single JWT access token with ES256 signing method.
func (jm *JWTManager) GenerateAccessToken(userID string, expiration time.Duration) (string, error) {
	if userID == "" {
		return "", errors.New("userID cannot be empty")
	}

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	return token.SignedString(jm.PrivateKey)
}

// ValidateToken verifies the JWT signature and returns the claims.
func (jm *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token string cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jm.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
