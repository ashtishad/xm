package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func generateTestKeys(tb testing.TB) ([]byte, []byte) {
	privateKey, publicKey, err := generateECDSAKeyPair()
	if err != nil {
		tb.Fatalf("Failed to generate test keys: %v", err)
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		tb.Fatalf("Failed to marshal private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		tb.Fatalf("Failed to marshal public key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKeyPEM, publicKeyPEM
}

func TestNewJWTManager(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)

	tests := []struct {
		name        string
		privateKey  []byte
		publicKey   []byte
		expectError bool
	}{
		{"Valid keys", privateKey, publicKey, false},
		{"Invalid private key", []byte("invalid"), publicKey, true},
		{"Invalid public key", privateKey, []byte("invalid"), true},
		{"Empty keys", []byte{}, []byte{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtManager, err := NewJWTManager(time.Millisecond*100, tt.privateKey, tt.publicKey)
			if tt.expectError && err == nil {
				t.Errorf("NewJWTManager() error = nil, expected an error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("NewJWTManager() unexpected error = %v", err)
			}
			if !tt.expectError && jwtManager == nil {
				t.Errorf("NewJWTManager() returned nil JWTManager")
			}
		})
	}
}

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	jwtManager, err := NewJWTManager(time.Millisecond*100, privateKey, publicKey)
	if err != nil {
		t.Errorf("Failed to create JWTManager: %v", err)
		return
	}

	tests := []struct {
		name        string
		userID      string
		expectError bool
	}{
		{"Valid user ID", "user123", false},
		{"Empty user ID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken, err := jwtManager.GenerateAccessToken(tt.userID, jwtManager.AccessExp)
			if tt.expectError && err == nil {
				t.Errorf("GenerateAccessToken() error = nil, expected an error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("GenerateAccessToken() unexpected error = %v", err)
			}
			if !tt.expectError && (accessToken == "") {
				t.Errorf("GenerateAccessToken() returned empty token(s)")
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	jwtManager, err := NewJWTManager(time.Second, privateKey, publicKey)
	if err != nil {
		t.Errorf("Failed to create JWTManager: %v", err)
		return
	}

	validToken, err := jwtManager.GenerateAccessToken("user123", time.Second*2)
	if err != nil {
		t.Errorf("Failed to generate valid token: %v", err)
		return
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectedID  string
	}{
		{"Valid token", validToken, false, "user123"},
		{"Empty token", "", true, ""},
		{"Invalid token", "invalid.token.here", true, ""},
		{"Malformed token", "invalid", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateToken() error = nil, expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateToken() unexpected error = %v", err)
				} else if claims == nil {
					t.Errorf("ValidateToken() returned nil claims")
				} else if claims.UserID != tt.expectedID {
					t.Errorf("ValidateToken() got UserID = %v, want %v", claims.UserID, tt.expectedID)
				}
			}
		})
	}
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	privateKey, publicKey := generateTestKeys(t)
	jwtManager, err := NewJWTManager(time.Millisecond*100, privateKey, publicKey)
	if err != nil {
		t.Errorf("Failed to create JWTManager: %v", err)
		return
	}

	veryShortDuration := time.Millisecond * 10

	token, err := jwtManager.GenerateAccessToken("user123", veryShortDuration)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(veryShortDuration)

	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Errorf("ValidateToken() error = nil, expected an error for expired token")
	}
}

func FuzzJWTManager_GenerateAndValidateToken(f *testing.F) {
	f.Add("user123")
	f.Add("")
	f.Add("a")
	f.Add(strings.Repeat("a", 1000))

	privateKey, publicKey := generateTestKeys(f)
	jwtManager, err := NewJWTManager(time.Millisecond*100, privateKey, publicKey)
	if err != nil {
		f.Errorf("Failed to create JWTManager: %v", err)
		return
	}

	f.Fuzz(func(t *testing.T, userID string) {
		accessToken, err := jwtManager.GenerateAccessToken(userID, jwtManager.AccessExp)
		if err != nil {
			return // Skip invalid inputs
		}

		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			return // Skip invalid tokens
		}

		if claims.UserID != userID {
			t.Errorf("ValidateToken() got UserID = %v, want %v", claims.UserID, userID)
		}
	})
}

// test helpers
// ParseECKeyFromEnv decodes esdsa private and public keys
func ParseECKeyFromEnv(keyBase64 string) (*ecdsa.PrivateKey, error) {
	if keyBase64 == "" {
		return nil, errors.New("key cannot be empty")
	}

	keyPEM, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// generateECDSAKeyPair creates a new ECDSA key pair for JWT signing and verification.
func generateECDSAKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}
