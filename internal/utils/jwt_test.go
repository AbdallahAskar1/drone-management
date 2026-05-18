package utils

import (
	"drone-management/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestJWTSigner(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour
	signer := NewJWTSigner(secret, ttl)
	now := time.Now().Truncate(time.Second)

	t.Run("Issue and Parse", func(t *testing.T) {
		token, err := signer.Issue(123, "Alice", domain.RoleAdmin, now)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := signer.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, "Alice", claims.Name)
		assert.Equal(t, domain.RoleAdmin, claims.Role)
		assert.Equal(t, "123", claims.Subject)

		pid, err := ClaimSubjectUint(claims)
		require.NoError(t, err)
		assert.Equal(t, uint(123), pid)
	})

	t.Run("Invalid Secret", func(t *testing.T) {
		token, _ := signer.Issue(123, "Alice", domain.RoleAdmin, now)
		badSigner := NewJWTSigner("wrong-secret", ttl)
		_, err := badSigner.Parse(token)
		assert.Error(t, err)
	})

	t.Run("Expired Token", func(t *testing.T) {
		token, _ := signer.Issue(123, "Alice", domain.RoleAdmin, now.Add(-2*time.Hour))
		_, err := signer.Parse(token)
		assert.Error(t, err)
	})
}

func TestUintToStr(t *testing.T) {
	assert.Equal(t, "123", UintToStr(123))
	assert.Equal(t, "0", UintToStr(0))
	assert.Equal(t, "999999", UintToStr(999999))
}
