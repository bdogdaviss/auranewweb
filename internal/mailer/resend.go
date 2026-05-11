package mailer

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/resend/resend-go/v2"
)

//go:embed license_email.html
var licenseEmailHTML string

// LicenseEmailTemplate returns the embedded transactional HTML for unit tests
// or manual inspection. The runtime substitution is done in renderLicenseHTML.
func LicenseEmailTemplate() string { return licenseEmailHTML }

type ResendMailer struct {
	client       *resend.Client
	from         string
	supportEmail string
	downloadURL  string
}

// NewResendMailer constructs a Mailer that sends through Resend.
//   - apiKey       — your `re_...` key
//   - from         — verified sender; e.g. `Aura Optimizer <licenses@mail.example.com>`
//   - supportEmail — shown in the email's "Need help?" line; falls back to the
//                    address portion of `from` if empty.
//   - downloadURL  — the "Download Aura Optimizer" button target; must be set,
//                    or the button in the email will be empty.
func NewResendMailer(apiKey, from, supportEmail, downloadURL string) (*ResendMailer, error) {
	if apiKey == "" {
		return nil, errors.New("resend: api key is required")
	}
	if from == "" {
		return nil, errors.New("resend: from address is required")
	}
	if supportEmail == "" {
		supportEmail = addressPart(from)
	}
	return &ResendMailer{
		client:       resend.NewClient(apiKey),
		from:         from,
		supportEmail: supportEmail,
		downloadURL:  downloadURL,
	}, nil
}

func (m *ResendMailer) SendLicense(ctx context.Context, to, licenseKey, productID string) error {
	if to == "" {
		return errors.New("recipient is required")
	}

	htmlBody := renderLicenseHTML(licenseEmailHTML, map[string]string{
		"LICENSE_KEY":    licenseKey,
		"SUPPORT_EMAIL":  m.supportEmail,
		"CUSTOMER_EMAIL": to,
		"DOWNLOAD_URL":   m.downloadURL,
	})

	textBody := fmt.Sprintf(
		"Thanks for your purchase!\n\n"+
			"Your %s license key:\n\n%s\n\n"+
			"Download Aura Optimizer: %s\n\n"+
			"Keep this key private — do not share it.\n\n"+
			"Need help? Reply to this email or contact %s.\n",
		productID, licenseKey, m.downloadURL, m.supportEmail,
	)

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: "Your Aura Optimizer license key",
		Html:    htmlBody,
		Text:    textBody,
		ReplyTo: m.supportEmail,
	}
	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	return nil
}

// renderLicenseHTML does dumb {{NAME}} substitution. Values are HTML-escaped
// to defang any user-controlled input (the recipient email comes from a Stripe
// PaymentIntent metadata field — Stripe doesn't sanitize it for us). License
// keys are alphanumeric+hyphen so escaping is a no-op for them, but cheap.
func renderLicenseHTML(tmpl string, vars map[string]string) string {
	for k, v := range vars {
		tmpl = strings.ReplaceAll(tmpl, "{{"+k+"}}", html.EscapeString(v))
	}
	return tmpl
}

// addressPart extracts the bare email out of a "Name <addr@host>" string.
// Returns the input unchanged if no <> wrapping is present.
func addressPart(s string) string {
	lt := strings.LastIndex(s, "<")
	gt := strings.LastIndex(s, ">")
	if lt >= 0 && gt > lt {
		return strings.TrimSpace(s[lt+1 : gt])
	}
	return strings.TrimSpace(s)
}
