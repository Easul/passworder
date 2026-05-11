package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionTimeout = 30 * time.Minute
	BCryptCost     = 12
)

type Auth struct {
	ID           int64  `db:"id" json:"id"`
	PasswordHash string `db:"password_hash" json:"-"`
	KDFSalt      []byte `db:"kdf_salt" json:"-"`
	CreatedAt    int64  `db:"created_at" json:"createdAt"`
	UpdatedAt    int64  `db:"updated_at" json:"updatedAt"`
}

type Service struct {
	passwordHash string
	salt         []byte
	session      *Session
	cryptoKey    []byte
}

type Session struct {
	Token     string
	CryptoKey []byte
	ExpiresAt time.Time
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) IsInitialized() bool {
	return s.passwordHash != ""
}

func (s *Service) Setup(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	if err != nil {
		return err
	}
	s.passwordHash = string(hash)

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	s.salt = salt

	// Derive crypto key from password and salt
	s.cryptoKey = argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	return nil
}

func (s *Service) Load(passwordHash string, salt []byte) {
	s.passwordHash = passwordHash
	s.salt = salt
}

func (s *Service) Verify(password string) error {
	if s.passwordHash == "" {
		return errors.New("not initialized")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(s.passwordHash), []byte(password)); err != nil {
		return err
	}
	// Derive crypto key on successful verification
	s.cryptoKey = argon2.IDKey([]byte(password), s.salt, 3, 64*1024, 4, 32)
	return nil
}

func (s *Service) CreateSession() (*Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	s.session = &Session{
		Token:     base64.URLEncoding.EncodeToString(tokenBytes),
		CryptoKey: s.cryptoKey,
		ExpiresAt: time.Now().Add(SessionTimeout),
	}

	return s.session, nil
}

func (s *Service) ValidateSession(token string) bool {
	if s.session == nil || s.session.Token != token {
		return false
	}
	if time.Now().After(s.session.ExpiresAt) {
		s.session = nil
		return false
	}
	s.session.ExpiresAt = time.Now().Add(SessionTimeout)
	if s.cryptoKey == nil && s.session.CryptoKey != nil {
		s.cryptoKey = s.session.CryptoKey
	}
	return true
}

func (s *Service) DestroySession() {
	s.session = nil
}

func (s *Service) GetSalt() []byte {
	return s.salt
}

func (s *Service) GetPasswordHash() string {
	return s.passwordHash
}

func (s *Service) GetCryptoKey() []byte {
	return s.cryptoKey
}

func (s *Service) HashPassword(password string) (hash string, salt []byte, err error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), BCryptCost)
	if err != nil {
		return "", nil, err
	}

	salt = make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", nil, err
	}

	return string(hashBytes), salt, nil
}
