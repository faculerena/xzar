package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"` // bcrypt hash
}

func LoadCredentials(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	if creds.Username == "" || creds.Password == "" {
		return nil, fmt.Errorf("credentials file must have username and password fields")
	}
	return &creds, nil
}

func (c *Credentials) Verify(username, password string) bool {
	if username != c.Username {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password)) == nil
}
