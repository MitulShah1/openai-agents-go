package session

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/openai/openai-go/v3"
)

const encryptedPrefix = "AGENTS_AES_GCM_BASE64:"

type encryptedSession struct {
	inner Session
	gcm   cipher.AEAD
}

// WithEncryption wraps a session backend with AES-GCM encryption.
// The key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
// Returns an error if the key is invalid or cipher block cannot be created.
func WithEncryption(s Session, key []byte) (Session, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &encryptedSession{
		inner: s,
		gcm:   gcm,
	}, nil
}

func (s *encryptedSession) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages, err := s.inner.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var result []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		content := getEncryptionMessageContent(msg)
		if strings.HasPrefix(content, encryptedPrefix) {
			decrypted, err := s.decryptMessages(content)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt message: %w", err)
			}
			result = append(result, decrypted...)
		} else {
			result = append(result, msg)
		}
	}

	return result, nil
}

func (s *encryptedSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	if len(messages) == 0 {
		return nil
	}

	encrypted, err := s.encryptMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to encrypt messages: %w", err)
	}

	blobMsg := openai.SystemMessage(encryptedPrefix + encrypted)
	return s.inner.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{blobMsg})
}

func (s *encryptedSession) Clear(ctx context.Context, sessionID string) error {
	return s.inner.Clear(ctx, sessionID)
}

func (s *encryptedSession) Delete(ctx context.Context, sessionID string) error {
	return s.inner.Delete(ctx, sessionID)
}

func (s *encryptedSession) encryptMessages(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := s.gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *encryptedSession) decryptMessages(content string) ([]openai.ChatCompletionMessageParamUnion, error) {
	encoded := strings.TrimPrefix(content, encryptedPrefix)

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var messages []openai.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(plaintext, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Copy of helper to avoid dependency on compression.go unexported func
func getEncryptionMessageContent(msg openai.ChatCompletionMessageParamUnion) string {
	data, err := json.Marshal(msg)
	if err != nil {
		return ""
	}

	if !strings.Contains(string(data), `"role":"system"`) {
		return ""
	}

	var partial struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &partial); err != nil {
		return ""
	}

	var contentStr string
	if err := json.Unmarshal(partial.Content, &contentStr); err == nil {
		return contentStr
	}

	return ""
}
