package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// tokenPayload is the JSON carried inside the token (the part after the dot).
// PHP builds it, Go reads it. Field names must match what PHP sends.
type tokenPayload struct {
	Usuario int   `json:"idusu_usuario"`
	Empresa int   `json:"idgen_empresa"`
	Exp     int64 `json:"exp"` // unix timestamp; 0 = no expiry
}

// ParseToken validates a "firma.payload" token and returns who it belongs to.
//
// Format (same scheme as PHP/JS ApiConnector):
//
//	token = base64url(HMAC_SHA256(payloadJSON, secret)) + "." + base64url(payloadJSON)
//
// base64url uses the -_ alphabet with no padding, matching PHP/JS.
func ParseToken(token, secret string) (usuario, empresa int, err error) {
	// 1) A valid token is exactly "signature.payload".
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, 0, errors.New("token mal formado: se esperaban 2 partes separadas por '.'")
	}
	sigB64, payloadB64 := parts[0], parts[1]

	// 2) Decode the payload back to the exact JSON bytes PHP signed.
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return 0, 0, fmt.Errorf("payload base64 inválido: %w", err)
	}

	// 3) Re-sign those bytes with our shared secret. If the token is legit,
	//    this must produce the exact same signature PHP put in the token.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadJSON)
	expectedSig := mac.Sum(nil)

	// 4) Decode the signature that arrived inside the token.
	gotSig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return 0, 0, fmt.Errorf("firma base64 inválida: %w", err)
	}

	// 5) Constant-time compare so an attacker can't guess the signature
	//    byte-by-byte by measuring how long the check takes.
	if !hmac.Equal(expectedSig, gotSig) {
		return 0, 0, errors.New("firma no coincide: token falso o secreto distinto")
	}

	// 6) Signature checks out -> we can trust the payload. Now read it.
	var p tokenPayload
	if err := json.Unmarshal(payloadJSON, &p); err != nil {
		return 0, 0, fmt.Errorf("payload JSON inválido: %w", err)
	}

	// 7) Reject expired tokens (exp == 0 means "never expires").
	if p.Exp != 0 && time.Now().Unix() > p.Exp {
		return 0, 0, errors.New("token expirado")
	}

	// 8) A token with no user/company is useless to us.
	if p.Usuario == 0 || p.Empresa == 0 {
		return 0, 0, errors.New("token sin idusu_usuario o idgen_empresa")
	}

	return p.Usuario, p.Empresa, nil
}