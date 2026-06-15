package repo

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// mustEnsureUser creates (or fetches) a durable user and fails the test on error.
func mustEnsureUser(t *testing.T, ur UserRepository, id, email string) *User {
	t.Helper()
	u, err := ur.EnsureUser(context.Background(), id, email, "")
	if err != nil {
		t.Fatalf("ensure user %s: %v", email, err)
	}
	if u == nil {
		t.Fatalf("ensure user %s returned nil", email)
	}
	return u
}

func TestEnsureUser_MintsCodeAndIsIdempotent(t *testing.T) {
	ur := NewUserRepository(mustOpen(t))

	first := mustEnsureUser(t, ur, "u1", "alice@example.com")
	if first.ReferralCode == "" {
		t.Fatal("expected a referral code to be minted")
	}
	if first.StoreCreditCents != 0 {
		t.Errorf("new user balance = %d, want 0", first.StoreCreditCents)
	}

	// Same email again: returns the stored row, same code, no new mint.
	again := mustEnsureUser(t, ur, "u-different-id", "alice@example.com")
	if again.ID != first.ID {
		t.Errorf("id changed for existing email: %q -> %q", first.ID, again.ID)
	}
	if again.ReferralCode != first.ReferralCode {
		t.Errorf("referral code changed on re-ensure: %q -> %q", first.ReferralCode, again.ReferralCode)
	}
}

func TestCreditReferral_HappyPath(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	referrer := mustEnsureUser(t, ur, "u1", "alice@example.com")

	credited, err := rr.CreditReferral(ctx, CreditReferralInput{
		StripeEventID: "evt_1",
		ReferralCode:  referrer.ReferralCode,
		BuyerEmail:    "bob@example.com",
		ProductID:     "lifetime",
		CreditCents:   599,
	})
	if err != nil {
		t.Fatalf("credit: %v", err)
	}
	if !credited {
		t.Fatal("expected credited=true")
	}

	if bal, _ := ur.GetBalance(ctx, referrer.ID); bal != 599 {
		t.Errorf("balance = %d, want 599", bal)
	}
	if n, _ := rr.CountByReferrer(ctx, referrer.ID); n != 1 {
		t.Errorf("referral count = %d, want 1", n)
	}
}

func TestCreditReferral_CodeCaseInsensitive(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	u := mustEnsureUser(t, ur, "u1", "alice@example.com") // code stored lowercase
	// Credit using an UPPERCASED code — must still match (regression guard for
	// the share-link case-mismatch bug).
	credited, err := rr.CreditReferral(ctx, CreditReferralInput{
		StripeEventID: "evt_case",
		ReferralCode:  strings.ToUpper(u.ReferralCode),
		BuyerEmail:    "bob@example.com",
		ProductID:     "lifetime",
		CreditCents:   599,
	})
	if err != nil || !credited {
		t.Fatalf("expected case-insensitive credit, got credited=%v err=%v", credited, err)
	}
	if bal, _ := ur.GetBalance(ctx, u.ID); bal != 599 {
		t.Errorf("balance = %d, want 599", bal)
	}
}

func TestCreditReferral_IdempotentPerEvent(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	referrer := mustEnsureUser(t, ur, "u1", "alice@example.com")
	in := CreditReferralInput{
		StripeEventID: "evt_dup",
		ReferralCode:  referrer.ReferralCode,
		BuyerEmail:    "bob@example.com",
		ProductID:     "lifetime",
		CreditCents:   599,
	}

	if credited, err := rr.CreditReferral(ctx, in); err != nil || !credited {
		t.Fatalf("first credit: credited=%v err=%v", credited, err)
	}
	// Webhook retry: same event id must not double-credit.
	if credited, err := rr.CreditReferral(ctx, in); err != nil {
		t.Fatalf("retry credit: %v", err)
	} else if credited {
		t.Error("retry should not credit again")
	}

	if bal, _ := ur.GetBalance(ctx, referrer.ID); bal != 599 {
		t.Errorf("balance = %d after retry, want 599", bal)
	}
	if n, _ := rr.CountByReferrer(ctx, referrer.ID); n != 1 {
		t.Errorf("referral count = %d after retry, want 1", n)
	}
}

func TestCreditReferral_BlocksSelfReferral(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	referrer := mustEnsureUser(t, ur, "u1", "alice@example.com")

	// Buyer email matches the referrer (case-insensitive) — no credit.
	credited, err := rr.CreditReferral(ctx, CreditReferralInput{
		StripeEventID: "evt_self",
		ReferralCode:  referrer.ReferralCode,
		BuyerEmail:    "Alice@Example.com",
		ProductID:     "lifetime",
		CreditCents:   599,
	})
	if err != nil {
		t.Fatalf("credit: %v", err)
	}
	if credited {
		t.Error("self-referral should not credit")
	}
	if bal, _ := ur.GetBalance(ctx, referrer.ID); bal != 0 {
		t.Errorf("balance = %d, want 0 (self-referral blocked)", bal)
	}
}

// fundReferrer credits the user `count` times (599¢ each) via distinct events.
func fundReferrer(t *testing.T, rr ReferralRepository, code string, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		if _, err := rr.CreditReferral(context.Background(), CreditReferralInput{
			StripeEventID: "fund_" + code + "_" + string(rune('a'+i)),
			ReferralCode:  code,
			BuyerEmail:    "buyer@example.com",
			ProductID:     "lifetime",
			CreditCents:   599,
		}); err != nil {
			t.Fatalf("fund: %v", err)
		}
	}
}

func TestRedeemCredit_DebitsBalance(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	u := mustEnsureUser(t, ur, "u1", "carol@example.com")
	fundReferrer(t, rr, u.ReferralCode, 2) // 1198¢
	if bal, _ := ur.GetBalance(ctx, u.ID); bal != 1198 {
		t.Fatalf("funded balance = %d, want 1198", bal)
	}

	redeemed, err := rr.RedeemCredit(ctx, RedeemInput{
		StripeEventID: "pe_1", Email: "carol@example.com", ProductID: "lifetime", AmountCents: 599,
	})
	if err != nil || !redeemed {
		t.Fatalf("redeem: redeemed=%v err=%v", redeemed, err)
	}
	if bal, _ := ur.GetBalance(ctx, u.ID); bal != 599 {
		t.Errorf("balance after redeem = %d, want 599", bal)
	}
}

func TestRedeemCredit_IdempotentPerEvent(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	u := mustEnsureUser(t, ur, "u1", "carol@example.com")
	fundReferrer(t, rr, u.ReferralCode, 2) // 1198¢
	in := RedeemInput{StripeEventID: "pe_dup", Email: "carol@example.com", ProductID: "lifetime", AmountCents: 599}

	if redeemed, err := rr.RedeemCredit(ctx, in); err != nil || !redeemed {
		t.Fatalf("first redeem: redeemed=%v err=%v", redeemed, err)
	}
	if redeemed, err := rr.RedeemCredit(ctx, in); err != nil {
		t.Fatalf("retry redeem: %v", err)
	} else if redeemed {
		t.Error("retry should not debit again")
	}
	if bal, _ := ur.GetBalance(ctx, u.ID); bal != 599 {
		t.Errorf("balance after retry = %d, want 599 (single debit)", bal)
	}
}

func TestRedeemCredit_ClampsAtZero(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()

	u := mustEnsureUser(t, ur, "u1", "carol@example.com")
	fundReferrer(t, rr, u.ReferralCode, 1) // 599¢

	// Debit more than the balance — must clamp at 0, never go negative.
	if _, err := rr.RedeemCredit(ctx, RedeemInput{
		StripeEventID: "pe_over", Email: "carol@example.com", ProductID: "lifetime", AmountCents: 999,
	}); err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if bal, _ := ur.GetBalance(ctx, u.ID); bal != 0 {
		t.Errorf("balance = %d, want 0 (clamped)", bal)
	}
}

func TestEnsureUserWithCode_PrefersUsername(t *testing.T) {
	ur := NewUserRepository(mustOpen(t))
	u, err := ur.EnsureUserWithCode(context.Background(), "u1", "dave@example.com", "Dave", "BdogBTW")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if u.ReferralCode != "bdogbtw" {
		t.Errorf("code = %q, want sanitized username 'bdogbtw'", u.ReferralCode)
	}
}

func TestEnsureUserWithCode_FallsBackWhenTaken(t *testing.T) {
	ur := NewUserRepository(mustOpen(t))
	// First user claims the username.
	first := mustEnsureUserCode(t, ur, "u1", "a@example.com", "bdogbtw")
	// Second user with the same desired code must fall back to a different code.
	second := mustEnsureUserCode(t, ur, "u2", "b@example.com", "bdogbtw")
	if second.ReferralCode == first.ReferralCode {
		t.Errorf("second user got the same code %q; expected a fallback", second.ReferralCode)
	}
}

func TestAddLead_DedupesAndValidates(t *testing.T) {
	lr := NewLeadRepository(mustOpen(t))
	ctx := context.Background()

	if err := lr.AddLead(ctx, "Lead@Example.com", "Lead", "affiliate-signup"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Same email (different case) + source → idempotent no-op, no error.
	if err := lr.AddLead(ctx, "lead@example.com", "Lead", "affiliate-signup"); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	// Empty email → error.
	if err := lr.AddLead(ctx, "", "x", "affiliate-signup"); err == nil {
		t.Error("expected error for empty email")
	}
}

func TestSetPasswordAndVerify(t *testing.T) {
	ur := NewUserRepository(mustOpen(t))
	ctx := context.Background()
	u := mustEnsureUser(t, ur, "u1", "dave@example.com")
	if u.PasswordHash != "" {
		t.Errorf("new user should have empty password hash, got %q", u.PasswordHash)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost)
	if err := ur.SetPassword(ctx, u.ID, string(hash)); err != nil {
		t.Fatalf("set password: %v", err)
	}

	got, _ := ur.FindByEmail(ctx, "dave@example.com")
	if got == nil || got.PasswordHash == "" {
		t.Fatal("password hash was not persisted")
	}
	if bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("hunter2")) != nil {
		t.Error("correct password should verify")
	}
	if bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("wrong")) == nil {
		t.Error("wrong password must NOT verify")
	}
}

func TestUpdateReferralCode(t *testing.T) {
	ur := NewUserRepository(mustOpen(t))
	ctx := context.Background()
	u := mustEnsureUser(t, ur, "u1", "dave@example.com")

	if err := ur.UpdateReferralCode(ctx, u.ID, "DaveStore_1"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := ur.FindByEmail(ctx, "dave@example.com")
	if got.ReferralCode != "davestore_1" {
		t.Errorf("code = %q, want 'davestore_1'", got.ReferralCode)
	}

	// Too-short code is rejected.
	if err := ur.UpdateReferralCode(ctx, u.ID, "ab"); err == nil {
		t.Error("expected error for too-short code")
	}

	// Taken by someone else is rejected.
	other := mustEnsureUser(t, ur, "u2", "erin@example.com")
	if err := ur.UpdateReferralCode(ctx, other.ID, "davestore_1"); err != ErrReferralCodeTaken {
		t.Errorf("err = %v, want ErrReferralCodeTaken", err)
	}
}

func TestListByReferrer_ReturnsEvents(t *testing.T) {
	db := mustOpen(t)
	ur := NewUserRepository(db)
	rr := NewReferralRepository(db)
	ctx := context.Background()
	u := mustEnsureUser(t, ur, "u1", "carol@example.com")
	fundReferrer(t, rr, u.ReferralCode, 2)

	events, err := rr.ListByReferrer(ctx, u.ID, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	if events[0].CreditCents != 599 || events[0].ProductID != "lifetime" {
		t.Errorf("unexpected event: %+v", events[0])
	}
	if total, _ := rr.TotalEarnedCents(ctx, u.ID); total != 1198 {
		t.Errorf("total earned = %d, want 1198", total)
	}
}

func mustEnsureUserCode(t *testing.T, ur UserRepository, id, email, code string) *User {
	t.Helper()
	u, err := ur.EnsureUserWithCode(context.Background(), id, email, "", code)
	if err != nil || u == nil {
		t.Fatalf("ensure %s: %v", email, err)
	}
	return u
}

func TestRedeemCredit_UnknownEmailIsNoOp(t *testing.T) {
	rr := NewReferralRepository(mustOpen(t))
	redeemed, err := rr.RedeemCredit(context.Background(), RedeemInput{
		StripeEventID: "pe_x", Email: "ghost@example.com", ProductID: "lifetime", AmountCents: 100,
	})
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if redeemed {
		t.Error("unknown email should not redeem")
	}
}

func TestCreditReferral_UnknownCodeIsNoOp(t *testing.T) {
	db := mustOpen(t)
	rr := NewReferralRepository(db)

	credited, err := rr.CreditReferral(context.Background(), CreditReferralInput{
		StripeEventID: "evt_x",
		ReferralCode:  "NOSUCHCODE",
		BuyerEmail:    "bob@example.com",
		ProductID:     "lifetime",
		CreditCents:   599,
	})
	if err != nil {
		t.Fatalf("credit: %v", err)
	}
	if credited {
		t.Error("unknown code should not credit")
	}
}
