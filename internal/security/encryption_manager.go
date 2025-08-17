package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// EncryptionManager handles data encryption and decryption
type EncryptionManager struct {
	logger       *logrus.Logger
	tracer       trace.Tracer
	config       EncryptionConfig
	masterKey    []byte
	keys         map[string]*EncryptionKey
	gcm          cipher.AEAD
	lastRotation time.Time
	mu           sync.RWMutex
	running      bool
	stopCh       chan struct{}
}

// EncryptionKey represents an encryption key
type EncryptionKey struct {
	ID        string    `json:"id"`
	Key       []byte    `json:"key"`
	Algorithm string    `json:"algorithm"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data      []byte    `json:"data"`
	KeyID     string    `json:"key_id"`
	Algorithm string    `json:"algorithm"`
	Nonce     []byte    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager(logger *logrus.Logger, config EncryptionConfig) (*EncryptionManager, error) {
	tracer := otel.Tracer("encryption-manager")

	em := &EncryptionManager{
		logger: logger,
		tracer: tracer,
		config: config,
		keys:   make(map[string]*EncryptionKey),
		stopCh: make(chan struct{}),
	}

	if config.Enabled {
		if err := em.initializeEncryption(); err != nil {
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}
	}

	return em, nil
}

// Start initializes the encryption manager
func (em *EncryptionManager) Start(ctx context.Context) error {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.Start")
	defer span.End()

	em.mu.Lock()
	defer em.mu.Unlock()

	if em.running {
		return fmt.Errorf("encryption manager is already running")
	}

	if !em.config.Enabled {
		em.logger.Info("Encryption manager is disabled")
		return nil
	}

	em.logger.Info("Starting encryption manager")

	// Start key rotation if enabled
	if em.config.KeyRotation > 0 {
		go em.rotateKeys()
	}

	em.running = true
	em.logger.Info("Encryption manager started successfully")

	return nil
}

// Stop shuts down the encryption manager
func (em *EncryptionManager) Stop(ctx context.Context) error {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.Stop")
	defer span.End()

	em.mu.Lock()
	defer em.mu.Unlock()

	if !em.running {
		return nil
	}

	em.logger.Info("Stopping encryption manager")

	close(em.stopCh)
	em.running = false
	em.logger.Info("Encryption manager stopped")

	return nil
}

// GetStatus returns the current encryption status
func (em *EncryptionManager) GetStatus(ctx context.Context) (*models.EncryptionStatus, error) {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.GetStatus")
	defer span.End()

	em.mu.RLock()
	defer em.mu.RUnlock()

	return &models.EncryptionStatus{
		Enabled:      em.config.Enabled,
		Algorithm:    em.config.Algorithm,
		KeySize:      em.config.KeySize,
		AtRest:       em.config.AtRest,
		InTransit:    em.config.InTransit,
		HSMEnabled:   em.config.HSMEnabled,
		LastRotation: em.lastRotation,
		Timestamp:    time.Now(),
	}, nil
}

// Encrypt encrypts data using the current encryption key
func (em *EncryptionManager) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.Encrypt")
	defer span.End()

	if !em.config.Enabled {
		return data, nil // Return unencrypted if encryption is disabled
	}

	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.gcm == nil {
		return nil, fmt.Errorf("encryption not initialized")
	}

	// Generate a random nonce
	nonce := make([]byte, em.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := em.gcm.Seal(nonce, nonce, data, nil)

	// Create encrypted data structure
	encryptedData := &EncryptedData{
		Data:      ciphertext,
		KeyID:     "current", // TODO: Use actual key ID
		Algorithm: em.config.Algorithm,
		Nonce:     nonce,
		CreatedAt: time.Now(),
	}

	// Serialize encrypted data
	return em.serializeEncryptedData(encryptedData)
}

// Decrypt decrypts data using the appropriate encryption key
func (em *EncryptionManager) Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error) {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.Decrypt")
	defer span.End()

	if !em.config.Enabled {
		return encryptedData, nil // Return as-is if encryption is disabled
	}

	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.gcm == nil {
		return nil, fmt.Errorf("encryption not initialized")
	}

	// Deserialize encrypted data
	data, err := em.deserializeEncryptedData(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize encrypted data: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := em.gcm.NonceSize()
	if len(data.Data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data.Data[:nonceSize], data.Data[nonceSize:]

	// Decrypt the data
	plaintext, err := em.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (em *EncryptionManager) EncryptString(ctx context.Context, plaintext string) (string, error) {
	encrypted, err := em.Encrypt(ctx, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 encoded string
func (em *EncryptionManager) DecryptString(ctx context.Context, encryptedText string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	decrypted, err := em.Decrypt(ctx, encrypted)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GenerateKey generates a new encryption key
func (em *EncryptionManager) GenerateKey(ctx context.Context) (*EncryptionKey, error) {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.GenerateKey")
	defer span.End()

	keySize := em.config.KeySize / 8 // Convert bits to bytes
	if keySize == 0 {
		keySize = 32 // Default to 256 bits
	}

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	keyID := generateKeyID()
	encKey := &EncryptionKey{
		ID:        keyID,
		Key:       key,
		Algorithm: em.config.Algorithm,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(em.config.KeyRotation),
		Active:    true,
	}

	em.mu.Lock()
	em.keys[keyID] = encKey
	em.mu.Unlock()

	em.logger.WithField("key_id", keyID).Info("New encryption key generated")

	return encKey, nil
}

// RotateKey rotates the current encryption key
func (em *EncryptionManager) RotateKey(ctx context.Context) error {
	ctx, span := em.tracer.Start(ctx, "encryptionManager.RotateKey")
	defer span.End()

	em.logger.Info("Rotating encryption key")

	// Generate new key
	newKey, err := em.GenerateKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Update master key and GCM
	em.mu.Lock()
	em.masterKey = newKey.Key
	em.lastRotation = time.Now()

	// Reinitialize GCM with new key
	if err := em.initializeGCM(); err != nil {
		em.mu.Unlock()
		return fmt.Errorf("failed to initialize GCM with new key: %w", err)
	}

	// Mark old keys as inactive
	for _, key := range em.keys {
		if key.ID != newKey.ID {
			key.Active = false
		}
	}

	em.mu.Unlock()

	em.logger.WithField("key_id", newKey.ID).Info("Encryption key rotated successfully")

	return nil
}

// Helper methods

func (em *EncryptionManager) initializeEncryption() error {
	// Generate or load master key
	if err := em.generateMasterKey(); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	// Initialize GCM
	if err := em.initializeGCM(); err != nil {
		return fmt.Errorf("failed to initialize GCM: %w", err)
	}

	em.lastRotation = time.Now()
	return nil
}

func (em *EncryptionManager) generateMasterKey() error {
	keySize := em.config.KeySize / 8 // Convert bits to bytes
	if keySize == 0 {
		keySize = 32 // Default to 256 bits
	}

	em.masterKey = make([]byte, keySize)
	if _, err := rand.Read(em.masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	return nil
}

func (em *EncryptionManager) initializeGCM() error {
	block, err := aes.NewCipher(em.masterKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	em.gcm = gcm
	return nil
}

func (em *EncryptionManager) serializeEncryptedData(data *EncryptedData) ([]byte, error) {
	// Simple serialization - in production, use proper serialization format
	return data.Data, nil
}

func (em *EncryptionManager) deserializeEncryptedData(data []byte) (*EncryptedData, error) {
	// Simple deserialization - in production, use proper deserialization format
	return &EncryptedData{
		Data:      data,
		KeyID:     "current",
		Algorithm: em.config.Algorithm,
		CreatedAt: time.Now(),
	}, nil
}

func (em *EncryptionManager) rotateKeys() {
	if em.config.KeyRotation <= 0 {
		return
	}

	ticker := time.NewTicker(em.config.KeyRotation)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := em.RotateKey(ctx); err != nil {
				em.logger.WithError(err).Error("Failed to rotate encryption key")
			}

		case <-em.stopCh:
			em.logger.Debug("Key rotation stopped")
			return
		}
	}
}

func generateKeyID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("key-%x", hash[:8])
}
