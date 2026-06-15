package mailer

import "context"

// Mailer sends transactional email. Concrete implementations live in this
// package (ResendMailer for production, NoopMailer for dev/test).
type Mailer interface {
	SendLicense(ctx context.Context, to, licenseKey, productID string) error
	// SendReferralEarned notifies a referrer that they earned store credit
	// from a successful referred purchase.
	SendReferralEarned(ctx context.Context, to, amount, buyerEmail, productID string) error
}
