package proofchain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// WebhookSignatureHeader is the header ProofChain sets on webhook deliveries.
const WebhookSignatureHeader = "X-Webhook-Signature"

// ErrInvalidWebhookSignature is returned when a webhook signature does not
// match the payload.
var ErrInvalidWebhookSignature = errors.New("proofchain: invalid webhook signature")

// VerifyWebhookSignature checks the X-Webhook-Signature header of a webhook
// delivery against the raw request body.
//
// ProofChain signs the exact bytes of the request body with HMAC-SHA256 using
// the webhook secret and sends the hex digest as "sha256=<hex>". Pass the raw
// body (do not re-serialise it). The comparison is constant-time.
//
// Deliveries carry no signed timestamp, so callers that need replay
// protection should de-duplicate on the X-Delivery-ID header.
func VerifyWebhookSignature(secret string, payload []byte, header string) error {
	if secret == "" {
		return errors.New("proofchain: webhook secret is required")
	}
	sig, ok := strings.CutPrefix(strings.TrimSpace(header), "sha256=")
	if !ok || sig == "" {
		return ErrInvalidWebhookSignature
	}
	got, err := hex.DecodeString(sig)
	if err != nil {
		return ErrInvalidWebhookSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return ErrInvalidWebhookSignature
	}
	return nil
}
