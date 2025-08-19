package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SecretsManager handles secure storage and retrieval of secrets
type SecretsManager struct {
	encryptionKey []byte
	secretsPath   string
}

// Secret represents a stored secret
type Secret struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(encryptionKey, secretsPath string) (*SecretsManager, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 characters long")
	}

	if secretsPath == "" {
		secretsPath = filepath.Join(os.Getenv("HOME"), ".aios", "secrets")
	}

	// Ensure secrets directory exists
	if err := os.MkdirAll(secretsPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}

	return &SecretsManager{
		encryptionKey: []byte(encryptionKey),
		secretsPath:   secretsPath,
	}, nil
}

// Store encrypts and stores a secret
func (sm *SecretsManager) Store(name, value, description string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	// Encrypt the value
	encryptedValue, err := sm.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Create secret file path
	secretFile := filepath.Join(sm.secretsPath, fmt.Sprintf("%s.secret", name))

	// Write encrypted value to file
	if err := os.WriteFile(secretFile, []byte(encryptedValue), 0600); err != nil {
		return fmt.Errorf("failed to write secret file: %w", err)
	}

	return nil
}

// Retrieve decrypts and retrieves a secret
func (sm *SecretsManager) Retrieve(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("secret name cannot be empty")
	}

	// Create secret file path
	secretFile := filepath.Join(sm.secretsPath, fmt.Sprintf("%s.secret", name))

	// Check if file exists
	if _, err := os.Stat(secretFile); os.IsNotExist(err) {
		return "", fmt.Errorf("secret '%s' not found", name)
	}

	// Read encrypted value from file
	encryptedValue, err := os.ReadFile(secretFile)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file: %w", err)
	}

	// Decrypt the value
	decryptedValue, err := sm.decrypt(string(encryptedValue))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return decryptedValue, nil
}

// Delete removes a secret
func (sm *SecretsManager) Delete(name string) error {
	if name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	// Create secret file path
	secretFile := filepath.Join(sm.secretsPath, fmt.Sprintf("%s.secret", name))

	// Check if file exists
	if _, err := os.Stat(secretFile); os.IsNotExist(err) {
		return fmt.Errorf("secret '%s' not found", name)
	}

	// Remove the file
	if err := os.Remove(secretFile); err != nil {
		return fmt.Errorf("failed to delete secret file: %w", err)
	}

	return nil
}

// List returns a list of all stored secret names
func (sm *SecretsManager) List() ([]string, error) {
	files, err := os.ReadDir(sm.secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets directory: %w", err)
	}

	var secrets []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".secret") {
			secretName := strings.TrimSuffix(file.Name(), ".secret")
			secrets = append(secrets, secretName)
		}
	}

	return secrets, nil
}

// Exists checks if a secret exists
func (sm *SecretsManager) Exists(name string) bool {
	secretFile := filepath.Join(sm.secretsPath, fmt.Sprintf("%s.secret", name))
	_, err := os.Stat(secretFile)
	return !os.IsNotExist(err)
}

// encrypt encrypts a plaintext string using AES-GCM
func (sm *SecretsManager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a base64-encoded ciphertext using AES-GCM
func (sm *SecretsManager) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// LoadSecretsFromEnv loads secrets from environment variables
func LoadSecretsFromEnv(prefix string) map[string]string {
	secrets := make(map[string]string)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key, value := pair[0], pair[1]
		if strings.HasPrefix(key, prefix) {
			secretName := strings.TrimPrefix(key, prefix)
			secretName = strings.ToLower(secretName)
			secrets[secretName] = value
		}
	}

	return secrets
}

// GenerateEncryptionKey generates a random 32-byte encryption key
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key)[:32], nil
}

// ValidateEncryptionKey validates that an encryption key is the correct length
func ValidateEncryptionKey(key string) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be exactly 32 characters long, got %d", len(key))
	}
	return nil
}

// SecretProvider defines an interface for secret providers
type SecretProvider interface {
	GetSecret(name string) (string, error)
	SetSecret(name, value string) error
	DeleteSecret(name string) error
	ListSecrets() ([]string, error)
}

// EnvironmentSecretProvider implements SecretProvider using environment variables
type EnvironmentSecretProvider struct {
	prefix string
}

// NewEnvironmentSecretProvider creates a new environment secret provider
func NewEnvironmentSecretProvider(prefix string) *EnvironmentSecretProvider {
	return &EnvironmentSecretProvider{
		prefix: prefix,
	}
}

// GetSecret retrieves a secret from environment variables
func (esp *EnvironmentSecretProvider) GetSecret(name string) (string, error) {
	envVar := fmt.Sprintf("%s_%s", esp.prefix, strings.ToUpper(name))
	value := os.Getenv(envVar)
	if value == "" {
		return "", fmt.Errorf("secret '%s' not found in environment", name)
	}
	return value, nil
}

// SetSecret sets a secret in environment variables (runtime only)
func (esp *EnvironmentSecretProvider) SetSecret(name, value string) error {
	envVar := fmt.Sprintf("%s_%s", esp.prefix, strings.ToUpper(name))
	return os.Setenv(envVar, value)
}

// DeleteSecret removes a secret from environment variables (runtime only)
func (esp *EnvironmentSecretProvider) DeleteSecret(name string) error {
	envVar := fmt.Sprintf("%s_%s", esp.prefix, strings.ToUpper(name))
	return os.Unsetenv(envVar)
}

// ListSecrets lists all secrets with the given prefix
func (esp *EnvironmentSecretProvider) ListSecrets() ([]string, error) {
	var secrets []string
	prefix := esp.prefix + "_"

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		if strings.HasPrefix(key, prefix) {
			secretName := strings.TrimPrefix(key, prefix)
			secretName = strings.ToLower(secretName)
			secrets = append(secrets, secretName)
		}
	}

	return secrets, nil
}

// FileSecretProvider implements SecretProvider using the SecretsManager
type FileSecretProvider struct {
	manager *SecretsManager
}

// NewFileSecretProvider creates a new file secret provider
func NewFileSecretProvider(encryptionKey, secretsPath string) (*FileSecretProvider, error) {
	manager, err := NewSecretsManager(encryptionKey, secretsPath)
	if err != nil {
		return nil, err
	}

	return &FileSecretProvider{
		manager: manager,
	}, nil
}

// GetSecret retrieves a secret from encrypted files
func (fsp *FileSecretProvider) GetSecret(name string) (string, error) {
	return fsp.manager.Retrieve(name)
}

// SetSecret stores a secret in encrypted files
func (fsp *FileSecretProvider) SetSecret(name, value string) error {
	return fsp.manager.Store(name, value, "")
}

// DeleteSecret removes a secret from encrypted files
func (fsp *FileSecretProvider) DeleteSecret(name string) error {
	return fsp.manager.Delete(name)
}

// ListSecrets lists all stored secrets
func (fsp *FileSecretProvider) ListSecrets() ([]string, error) {
	return fsp.manager.List()
}
