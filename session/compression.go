package session

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/openai/openai-go/v3"
)

const compressedPrefix = "AGENTS_GZIP_BASE64:"

type compressedSession struct {
	inner Session
}

// WithCompression wraps a session backend with transparent compression.
// Access to the underlying storage will see compressed data blobs stored as System messages.
func WithCompression(s Session) Session {
	return &compressedSession{inner: s}
}

func (s *compressedSession) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages, err := s.inner.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var result []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		// Check if it's a compressed message
		content := getMessageContent(msg)
		if strings.HasPrefix(content, compressedPrefix) {
			decompressed, err := decompressMessages(content)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress message: %w", err)
			}
			result = append(result, decompressed...)
		} else {
			result = append(result, msg)
		}
	}

	return result, nil
}

func (s *compressedSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	if len(messages) == 0 {
		return nil
	}

	// Compress the batch of messages
	compressed, err := compressMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to compress messages: %w", err)
	}

	// Create a single system message containing the blob
	blobMsg := openai.SystemMessage(compressedPrefix + compressed)

	return s.inner.Append(ctx, sessionID, []openai.ChatCompletionMessageParamUnion{blobMsg})
}

func (s *compressedSession) Clear(ctx context.Context, sessionID string) error {
	return s.inner.Clear(ctx, sessionID)
}

func (s *compressedSession) Delete(ctx context.Context, sessionID string) error {
	return s.inner.Delete(ctx, sessionID)
}

// Helper to extract string content from a message if possible
func getMessageContent(msg openai.ChatCompletionMessageParamUnion) string {
	// Attempt to marshal to JSON to inspect if we can't access fields directly
	// But first, let's try assuming it works like the Union struct in recent SDKs
	// which might be marshaled.
	// Actually, the most robust way without knowing internal struct fields is JSON roundtrip
	// OR seeing if the struct exposes "System" accessor.

	// Since we can't see the type definition, let's try generic JSON inspection approach for now
	// which matches how we store it (via JSON marshaling). It's slower but safe.
	data, err := json.Marshal(msg)
	if err != nil {
		return ""
	}

	// Fast check for role="system" in JSON
	// {"role":"system","content":"..."}
	if !strings.Contains(string(data), `"role":"system"`) {
		return ""
	}

	// Parse partially
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

func compressMessages(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return "", err
	}
	if err := gw.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func decompressMessages(content string) ([]openai.ChatCompletionMessageParamUnion, error) {
	encoded := strings.TrimPrefix(content, compressedPrefix)

	compressedData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gr.Close() }()

	var decompressedData bytes.Buffer
	// Limit decompression size to 50MB to prevent decompression bombs (G110)
	const maxDecompressionSize = 50 * 1024 * 1024
	if _, err := io.Copy(&decompressedData, io.LimitReader(gr, maxDecompressionSize)); err != nil {
		return nil, err
	}

	var messages []openai.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(decompressedData.Bytes(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
