package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	cookieName = "cms_session"
	maxAge     = 7 * 24 * time.Hour
)

// Store manages signed session cookies.
type Store struct {
	secret []byte
	secure bool
}

// Data holds session values.
type Data struct {
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewStore creates a session store with the given secret.
func NewStore(secret string, secure bool) *Store {
	return &Store{secret: []byte(secret), secure: secure}
}

// Get reads and validates the session cookie from the request.
func (s *Store) Get(r *http.Request) (*Data, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	payload, err := s.verify(cookie.Value)
	if err != nil {
		return nil, err
	}

	var data Data
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	if time.Now().After(data.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return &data, nil
}

// Set writes a signed session cookie to the response.
func (s *Store) Set(w http.ResponseWriter, data *Data) error {
	data.ExpiresAt = time.Now().Add(maxAge)

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	value, err := s.sign(payload)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secure,
	})

	return nil
}

// Clear removes the session cookie.
func (s *Store) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secure,
	})
}

func (s *Store) sign(payload []byte) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(nonce)
	mac.Write(payload)
	sig := mac.Sum(nil)

	combined := append(nonce, payload...)
	combined = append(combined, sig...)

	return base64.RawURLEncoding.EncodeToString(combined), nil
}

func (s *Store) verify(value string) ([]byte, error) {
	combined, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	if len(combined) < 16+sha256.Size {
		return nil, errors.New("invalid session")
	}

	nonce := combined[:16]
	payload := combined[16 : len(combined)-sha256.Size]
	sig := combined[len(combined)-sha256.Size:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(nonce)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return nil, errors.New("invalid session signature")
	}

	return payload, nil
}

// IsAuthenticated checks whether the request has a valid session.
func (s *Store) IsAuthenticated(r *http.Request) bool {
	data, err := s.Get(r)
	return err == nil && data != nil && strings.TrimSpace(data.Username) != ""
}
