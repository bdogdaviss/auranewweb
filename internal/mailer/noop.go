package mailer

import (
	"context"
	"log"
)

// NoopMailer logs the would-be send and returns nil. Used at boot when no
// email provider is configured, so the rest of the system stays operational
// (licenses still get persisted; the user can ask for a resend once email
// is wired up).
type NoopMailer struct{}

func (NoopMailer) SendLicense(_ context.Context, to, licenseKey, productID string) error {
	log.Printf("noop mailer: would send license %s for %s to %s", licenseKey, productID, to)
	return nil
}

func (NoopMailer) SendReferralEarned(_ context.Context, to, amount, buyerEmail, productID string) error {
	log.Printf("noop mailer: would send referral earned %s for %s (buyer %s) to %s", amount, productID, buyerEmail, to)
	return nil
}
