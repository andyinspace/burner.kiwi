package burner

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// injectWelcomeMessage composes and saves a welcome email into the given inbox.
// It reads WEBSITE_URL and DONATION_* environment variables to build the content.
func (s *Server) injectWelcomeMessage(i Inbox) error {
	subject := fmt.Sprintf("Welcome to %s", s.cfg.URL)

	body := composeWelcomeBody(s.cfg.URL)

	// Derive a reasonable from address from the website URL
	fromName, fromAddr := deriveFromIdentity(s.cfg.URL)

	msg := Message{
		ID:          uuid.Must(uuid.NewRandom()).String(),
		InboxID:     i.ID,
		TTL:         i.TTL,
		ReceivedAt:  time.Now().Unix(),
		Sender:      fromAddr,
		FromName:    fromName,
		FromAddress: fromAddr,
		Subject:     subject,
		BodyPlain:   body,
	}

	return s.db.SaveNewMessage(msg)
}

// composeWelcomeBody builds the plain text welcome email body using WEBSITE_URL and DONATION_* env vars.
func composeWelcomeBody(websiteURL string) string {
	// Collect donation addresses from env, include only non-empty values
	type crypto struct {
		Name    string
		EnvName string
	}

	cryptos := []crypto{
		{Name: "Bitcoin", EnvName: "DONATION_BITCOIN"},
		{Name: "Dash", EnvName: "DONATION_DASH"},
		{Name: "Doge", EnvName: "DONATION_DOGE"},
		{Name: "Ethereum", EnvName: "DONATION_ETHEREUM"},
		{Name: "Litecoin", EnvName: "DONATION_LITECOIN"},
		{Name: "Monero", EnvName: "DONATION_MONERO"},
		{Name: "Tron", EnvName: "DONATION_TRON"},
		{Name: "Zcash", EnvName: "DONATION_ZCASH"},
	}

	type pair struct {
		Name    string
		Address string
	}

	var available []pair
	for _, c := range cryptos {
		if v := strings.TrimSpace(os.Getenv(c.EnvName)); v != "" {
			available = append(available, pair{Name: c.Name, Address: v})
		}
	}

	sort.Slice(available, func(i, j int) bool { return available[i].Name < available[j].Name })

	var b strings.Builder
	fmt.Fprintf(&b, "Welcome to %s\n\n", websiteURL)
	b.WriteString("Your new disposable inbox is active and will expire in 23 hours and 59 minutes from its creation. To extend, you can click on the Extend icon.\n\n")
	b.WriteString("Note down the email address and your account recovery password. Without these, you cannot return to this mailbox.\n\n")
	b.WriteString("Keeping this service online costs real money—domains, infrastructure, bandwidth, and maintenance. If you find it useful, please consider donating a little crypto. Even a few dollars helps keep the service running and ad-free.\n\n")

	if len(available) > 0 {
		b.WriteString("Supported crypto:\n")
		for _, p := range available {
			fmt.Fprintf(&b, "- %s: %s\n", p.Name, p.Address)
		}
	} else {
		// If no donation methods are configured, omit the list but keep the nudge
		// No additional content
	}

	return b.String()
}

// deriveFromIdentity produces a human-friendly From name and address from WEBSITE_URL.
func deriveFromIdentity(websiteURL string) (string, string) {
	// Default fallback
	fallbackHost := "burner.kiwi"
	if websiteURL == "" {
		return "Welcome", "welcome@" + fallbackHost
	}

	u, err := url.Parse(websiteURL)
	if err != nil || u.Host == "" {
		// Treat websiteURL as host if parse fails
		host := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(websiteURL, "https://"), "http://"))
		if host == "" {
			host = fallbackHost
		}
		return host, "welcome@" + host
	}

	host := u.Host
	// Strip port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host, "welcome@" + host
}
