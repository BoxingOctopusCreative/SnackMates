package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebAuthnService struct {
	wa *webauthn.WebAuthn
}

func NewWebAuthnService(cfg config.Config) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.WebAuthnRPName,
		RPID:          cfg.WebAuthnRPID,
		RPOrigins:     []string{cfg.WebAuthnOrigin},
	}
	wa, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}
	return &WebAuthnService{wa: wa}, nil
}

type webAuthnUser struct {
	id          uuid.UUID
	name        string
	credentials []webauthn.Credential
}

func (u webAuthnUser) WebAuthnID() []byte                         { return u.id[:] }
func (u webAuthnUser) WebAuthnName() string                         { return u.name }
func (u webAuthnUser) WebAuthnDisplayName() string                { return u.name }
func (u webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func loadWebAuthnUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*webAuthnUser, error) {
	rec, err := GetUserByID(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	rows, err := pool.Query(ctx, `
		SELECT credential_id, public_key, sign_count FROM webauthn_credentials WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var credID, pubKey []byte
		var signCount uint32
		if err := rows.Scan(&credID, &pubKey, &signCount); err != nil {
			return nil, err
		}
		creds = append(creds, webauthn.Credential{
			ID:              credID,
			PublicKey:       pubKey,
			Authenticator:   webauthn.Authenticator{SignCount: signCount},
			Transport:       []protocol.AuthenticatorTransport{protocol.Internal, protocol.USB, protocol.NFC, protocol.BLE},
			AttestationType: "none",
		})
	}
	return &webAuthnUser{id: rec.ID, name: rec.Email, credentials: creds}, nil
}

type RegistrationBeginResponse struct {
	Options     *protocol.CredentialCreation `json:"options"`
	SessionData string                       `json:"session_data"`
}

func (s *WebAuthnService) BeginRegistration(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*RegistrationBeginResponse, error) {
	user, err := loadWebAuthnUser(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	options, sessionData, err := s.wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}
	encoded, err := encodeSession(sessionData)
	if err != nil {
		return nil, err
	}
	return &RegistrationBeginResponse{Options: options, SessionData: encoded}, nil
}

func (s *WebAuthnService) FinishRegistration(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, sessionEncoded string, response *protocol.ParsedCredentialCreationData, deviceName string) error {
	user, err := loadWebAuthnUser(ctx, pool, userID)
	if err != nil {
		return err
	}
	sessionData, err := decodeSession[webauthn.SessionData](sessionEncoded)
	if err != nil {
		return err
	}
	credential, err := s.wa.CreateCredential(user, *sessionData, response)
	if err != nil {
		return err
	}
	if deviceName == "" {
		deviceName = "Security Key"
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO webauthn_credentials (user_id, credential_id, public_key, sign_count, device_name)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, credential.ID, credential.PublicKey, credential.Authenticator.SignCount, deviceName)
	return err
}

type LoginBeginResponse struct {
	Options     *protocol.CredentialAssertion `json:"options"`
	SessionData string                        `json:"session_data"`
}

func (s *WebAuthnService) BeginLogin(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*LoginBeginResponse, error) {
	user, err := loadWebAuthnUser(ctx, pool, userID)
	if err != nil {
		return nil, err
	}
	if len(user.credentials) == 0 {
		return nil, fmt.Errorf("no webauthn credentials registered")
	}
	options, sessionData, err := s.wa.BeginLogin(user)
	if err != nil {
		return nil, err
	}
	encoded, err := encodeSession(sessionData)
	if err != nil {
		return nil, err
	}
	return &LoginBeginResponse{Options: options, SessionData: encoded}, nil
}

func (s *WebAuthnService) FinishLogin(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, sessionEncoded string, response *protocol.ParsedCredentialAssertionData) error {
	user, err := loadWebAuthnUser(ctx, pool, userID)
	if err != nil {
		return err
	}
	sessionData, err := decodeSession[webauthn.SessionData](sessionEncoded)
	if err != nil {
		return err
	}
	credential, err := s.wa.ValidateLogin(user, *sessionData, response)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		UPDATE webauthn_credentials SET sign_count = $3
		WHERE user_id = $1 AND credential_id = $2
	`, userID, credential.ID, credential.Authenticator.SignCount)
	return err
}

func encodeSession(data *webauthn.SessionData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeSession[T any](encoded string) (*T, error) {
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var session T
	if err := json.Unmarshal(b, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
