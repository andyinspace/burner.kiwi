package burner

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestComposeWelcomeBody_SortingFiltering(t *testing.T) {
	// Set env vars
	os.Setenv("DONATION_BITCOIN", "bc1qbitcoin")
	os.Setenv("DONATION_ETHEREUM", "0xethereum")
	os.Setenv("DONATION_ZCASH", "t1zcash")
	os.Setenv("DONATION_MONERO", "") // should be filtered
	t.Cleanup(func() {
		os.Unsetenv("DONATION_BITCOIN")
		os.Unsetenv("DONATION_ETHEREUM")
		os.Unsetenv("DONATION_ZCASH")
		os.Unsetenv("DONATION_MONERO")
	})

	body := composeWelcomeBody("https://example.com")

	// Should include site URL heading
	assert.Contains(t, body, "Welcome to https://example.com")

	// Should include Supported crypto header
	assert.Contains(t, body, "Supported crypto:")

	// Ensure only configured coins appear
	assert.Contains(t, body, "Bitcoin: bc1qbitcoin")
	assert.Contains(t, body, "Ethereum: 0xethereum")
	assert.Contains(t, body, "Zcash: t1zcash")
	assert.NotContains(t, body, "Monero:")

	// Ensure A-Z ordering by name: Bitcoin, Ethereum, Zcash
	idxB := strings.Index(body, "Bitcoin: bc1qbitcoin")
	idxE := strings.Index(body, "Ethereum: 0xethereum")
	idxZ := strings.Index(body, "Zcash: t1zcash")
	if !(idxB < idxE && idxE < idxZ) {
		t.Fatalf("crypto list not in A-Z order: indexes B=%d E=%d Z=%d", idxB, idxE, idxZ)
	}
}

func TestServer_InjectWelcomeMessage_Saves(t *testing.T) {
	os.Setenv("DONATION_BITCOIN", "bc1qbitcoin")
	t.Cleanup(func() { os.Unsetenv("DONATION_BITCOIN") })

	mdb := new(MockDatabase)

	s := &Server{cfg: Config{URL: "https://example.com"}, db: mdb}

	inbox := Inbox{ID: "inbox-1", TTL: 1234567890}

	// Expect a SaveNewMessage with matching InboxID and TTL; subject must match
	mdb.On("SaveNewMessage", mock.MatchedBy(func(msg Message) bool {
		return msg.InboxID == inbox.ID &&
			msg.TTL == inbox.TTL &&
			strings.HasPrefix(msg.Subject, "Welcome to ") &&
			strings.Contains(msg.BodyPlain, "Supported crypto:")
	})).Return(nil)

	err := s.injectWelcomeMessage(inbox)
	assert.NoError(t, err)

	mdb.AssertExpectations(t)
}
