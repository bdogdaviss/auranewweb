package mailer

import (
	"context"
	"testing"
)

func TestNoopMailer_SendLicense(t *testing.T) {
	var m Mailer = NoopMailer{}
	if err := m.SendLicense(context.Background(), "u@example.com", "key-123", "lifetime"); err != nil {
		t.Errorf("NoopMailer.SendLicense returned error: %v", err)
	}
}

func TestNewResendMailer_RejectsEmptyKey(t *testing.T) {
	if _, err := NewResendMailer("", "from@example.com"); err == nil {
		t.Error("expected error for empty api key")
	}
}

func TestNewResendMailer_RejectsEmptyFrom(t *testing.T) {
	if _, err := NewResendMailer("re_test", ""); err == nil {
		t.Error("expected error for empty from")
	}
}

func TestResendMailer_SendLicense_RejectsEmptyRecipient(t *testing.T) {
	m, err := NewResendMailer("re_test", "from@example.com")
	if err != nil {
		t.Fatalf("constructor: %v", err)
	}
	if err := m.SendLicense(context.Background(), "", "key", "product"); err == nil {
		t.Error("expected error for empty recipient")
	}
}
