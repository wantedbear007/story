package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params defines the parameters for Argon2id key derivation.
// Argon2id is the recommended variant as it provides resistance
// against both side-channel and GPU-based attacks.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams provides a balanced security profile.
// These values should be adjusted based on production hardware capabilities.
// Current values target ~100ms verification time on modern hardware.
var DefaultParams = Argon2Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// PasswordHasher implements application-layer password hashing using Argon2id.
// The encoded format includes all parameters, enabling future parameter rotation
// without breaking existing hashes.
type PasswordHasher struct {
	params Argon2Params
}

func NewPasswordHasher(params ...Argon2Params) *PasswordHasher {
	p := DefaultParams
	if len(params) > 0 {
		p = params[0]
	}
	return &PasswordHasher{params: p}
}

// Hash generates an Argon2id hash and returns it in encoded format:
// $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
func (ph *PasswordHasher) Hash(password string) (string, error) {
	salt, err := generateRandomBytes(ph.params.SaltLength)
	if err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		ph.params.Iterations,
		ph.params.Memory,
		ph.params.Parallelism,
		ph.params.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		ph.params.Memory,
		ph.params.Iterations,
		ph.params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// Verify compares a password against an encoded Argon2id hash.
// Uses constant-time comparison for the hash portion to prevent timing attacks.
func (ph *PasswordHasher) Verify(password, encodedHash string) error {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return fmt.Errorf("decoding hash: %w", err)
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, computedHash) != 1 {
		return fmt.Errorf("password does not match hash")
	}

	return nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("reading random bytes: %w", err)
	}
	return b, nil
}

// decodeHash parses an encoded Argon2id hash string into its components.
func decodeHash(encodedHash string) (Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("parsing version: %w", err)
	}
	if version != 19 {
		return Argon2Params{}, nil, nil, fmt.Errorf("unsupported version: %d", version)
	}

	var params Argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("parsing params: %w", err)
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decoding salt: %w", err)
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decoding hash: %w", err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}
