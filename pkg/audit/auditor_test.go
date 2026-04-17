package audit

import (
	"crypto/ed25519"
	"testing"

	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAuditorSigning(t *testing.T) {
	auditor, err := NewAuditor()
	assert.NoError(t, err)

	result := &models.TaskResult{
		Data: map[string]interface{}{"answer": 42},
	}

	// 1. Record steps
	auditor.Record("INIT", "Started", nil)
	auditor.Record("FINISH", "Ended", nil)

	// 2. Sign
	sig, err := auditor.Sign(result)
	assert.NoError(t, err)
	assert.NotNil(t, sig)
	assert.Len(t, result.AuditTrail, 2)

	// 3. Verify
	ok, err := VerifyDerivation(auditor.pubKey, result)
	assert.NoError(t, err)
	assert.True(t, ok)

	// 4. Tamper
	result.Data["answer"] = 43
	ok, _ = VerifyDerivation(auditor.pubKey, result)
	assert.False(t, ok, "Signature should fail after data tampering")
}

func TestVerifyDerivation(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	result := &models.TaskResult{
		Data: map[string]interface{}{"foo": "bar"},
	}
	
	sig := ed25519.Sign(priv, []byte(`{"foo":"bar"}`))
	result.Signature = sig

	ok, _ := VerifyDerivation(pub, result)
	assert.True(t, ok)
}
