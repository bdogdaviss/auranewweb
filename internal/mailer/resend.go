package mailer

import (
	"context"
	"errors"
	"fmt"
	"html"

	"github.com/resend/resend-go/v2"
)

type ResendMailer struct {
	client *resend.Client
	from   string
}

// NewResendMailer returns a Mailer backed by the Resend transactional API.
// apiKey must be a `re_...` key; from must be a verified sender on the
// Resend account (e.g. `Aura Optimizer <licenses@your-domain.com>`).
func NewResendMailer(apiKey, from string) (*ResendMailer, error) {
	if apiKey == "" {
		return nil, errors.New("resend: api key is required")
	}
	if from == "" {
		return nil, errors.New("resend: from address is required")
	}
	return &ResendMailer{
		client: resend.NewClient(apiKey),
		from:   from,
	}, nil
}

func (m *ResendMailer) SendLicense(ctx context.Context, to, licenseKey, productID string) error {
	if to == "" {
		return errors.New("recipient is required")
	}

	escKey := html.EscapeString(licenseKey)
	escProduct := html.EscapeString(productID)

	htmlBody := fmt.Sprintf(`<p>Thanks for your purchase!</p>
<p>Your <strong>%s</strong> license key:</p>
<p style="font-family:monospace;font-size:16px;padding:12px;background:#f4f4f4;border-radius:4px;word-break:break-all;">%s</p>
<p>Keep this email; you can also re-request the key from your account page if you lose it.</p>`, escProduct, escKey)

	textBody := fmt.Sprintf("Thanks for your purchase!\n\nYour %s license key:\n\n%s\n\nKeep this email; you can re-request the key from your account page if you lose it.\n", productID, licenseKey)

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: "Your Aura Optimizer license key",
		Html:    htmlBody,
		Text:    textBody,
	}
	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	return nil
}
