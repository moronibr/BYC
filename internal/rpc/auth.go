package rpc

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
)

// User represents an RPC user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Created  int64  `json:"created"`
}

// AuthManager manages RPC authentication
type AuthManager struct {
	users      map[string]*User
	mu         sync.RWMutex
	authFile   string
	saltSize   int
	workFactor int
}

// NewAuthManager creates a new auth manager
func NewAuthManager(authFile string) *AuthManager {
	return &AuthManager{
		users:      make(map[string]*User),
		authFile:   authFile,
		saltSize:   16,
		workFactor: 12,
	}
}

// LoadUsers loads users from the auth file
func (am *AuthManager) LoadUsers() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(am.authFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %v", err)
	}

	// Read auth file
	data, err := os.ReadFile(am.authFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty auth file
			return am.SaveUsers()
		}
		return fmt.Errorf("failed to read auth file: %v", err)
	}

	// Parse users
	if err := json.Unmarshal(data, &am.users); err != nil {
		return fmt.Errorf("failed to parse auth file: %v", err)
	}

	return nil
}

// SaveUsers saves users to the auth file
func (am *AuthManager) SaveUsers() error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Marshal users
	data, err := json.MarshalIndent(am.users, "", "   ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %v", err)
	}

	// Write to file
	if err := os.WriteFile(am.authFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %v", err)
	}

	return nil
}

// CreateUser creates a new user
func (am *AuthManager) CreateUser(username, password, role string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if user exists
	if _, exists := am.users[username]; exists {
		return ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), am.workFactor)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	am.users[username] = &User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
		Created:  time.Now().Unix(),
	}

	// Save users
	return am.SaveUsers()
}

// Authenticate authenticates a user
func (am *AuthManager) Authenticate(username, password string) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Get user
	user, exists := am.users[username]
	if !exists {
		return ErrUserNotFound
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

// DeleteUser deletes a user
func (am *AuthManager) DeleteUser(username string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if user exists
	if _, exists := am.users[username]; !exists {
		return ErrUserNotFound
	}

	// Delete user
	delete(am.users, username)

	// Save users
	return am.SaveUsers()
}

// UpdateUser updates a user
func (am *AuthManager) UpdateUser(username, password, role string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if user exists
	user, exists := am.users[username]
	if !exists {
		return ErrUserNotFound
	}

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), am.workFactor)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}
		user.Password = string(hashedPassword)
	}

	// Update role if provided
	if role != "" {
		user.Role = role
	}

	// Save users
	return am.SaveUsers()
}

// GetUser gets a user
func (am *AuthManager) GetUser(username string) (*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Get user
	user, exists := am.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return copy of user
	return &User{
		Username: user.Username,
		Role:     user.Role,
		Created:  user.Created,
	}, nil
}

// ListUsers lists all users
func (am *AuthManager) ListUsers() []*User {
	am.mu.RLock()
	defer am.mu.RUnlock()

	users := make([]*User, 0, len(am.users))
	for _, user := range am.users {
		users = append(users, &User{
			Username: user.Username,
			Role:     user.Role,
			Created:  user.Created,
		})
	}
	return users
}

// GenerateToken generates a new authentication token
func (am *AuthManager) GenerateToken(username string) (string, error) {
	// Generate random bytes
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	// Encode token
	return base64.URLEncoding.EncodeToString(token), nil
}

// ValidateToken validates an authentication token
func (am *AuthManager) ValidateToken(token string) bool {
	// Decode token
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	// Check token length
	if len(decoded) != 32 {
		return false
	}

	// TODO: Implement token validation logic
	return true
}
