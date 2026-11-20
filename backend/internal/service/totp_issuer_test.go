//go:build unit

package service

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// totpIssuerCacheStub 只实现 InitiateSetup 用到的 SetSetupSession；其余方法调用即 panic。
type totpIssuerCacheStub struct {
	TotpCache
	session *TotpSetupSession
}

func (s *totpIssuerCacheStub) SetSetupSession(ctx context.Context, userID int64, session *TotpSetupSession, ttl time.Duration) error {
	s.session = session
	return nil
}

func initiateTotpSetupWithSettings(t *testing.T, values map[string]string) *url.URL {
	t.Helper()
	admin := &User{ID: 7, Email: "admin@example.com", Role: RoleAdmin}
	require.NoError(t, admin.SetPassword("correct-horse"))

	values[SettingKeyTotpEnabled] = "true"
	settingSvc := NewSettingService(&totpVMSettingRepoStub{values: values}, nil)
	svc := NewTotpService(&totpVMUserRepoStub{user: admin}, nil, &totpIssuerCacheStub{}, settingSvc, nil, nil)

	resp, err := svc.InitiateSetup(context.Background(), admin.ID, "", "correct-horse")
	require.NoError(t, err)

	parsed, err := url.Parse(resp.QRCodeURL)
	require.NoError(t, err)
	return parsed
}

func TestTotpInitiateSetupUsesConfiguredSiteNameAsIssuer(t *testing.T) {
	parsed := initiateTotpSetupWithSettings(t, map[string]string{SettingKeySiteName: "  goreach.es  "})

	require.Equal(t, "goreach.es", parsed.Query().Get("issuer"))
	require.Equal(t, "/goreach.es:admin@example.com", parsed.Path)
	require.NotContains(t, parsed.String(), "Habifying")
}

func TestTotpInitiateSetupFallsBackToDefaultSiteNameIssuer(t *testing.T) {
	parsed := initiateTotpSetupWithSettings(t, map[string]string{})

	require.Equal(t, DefaultSiteName, parsed.Query().Get("issuer"))
	require.Equal(t, "Sub2API", parsed.Query().Get("issuer"))
}

func TestTotpIssuerWithoutSettingServiceUsesDefault(t *testing.T) {
	svc := &TotpService{}
	require.Equal(t, DefaultSiteName, svc.totpIssuer(context.Background()))
}
