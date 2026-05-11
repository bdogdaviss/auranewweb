package mailer

import (
	"context"
	"strings"
	"testing"
)

func TestNoopMailer_SendLicense(t *testing.T) {
	var m Mailer = NoopMailer{}
	if err := m.SendLicense(context.Background(), "u@example.com", "key-123", "lifetime"); err != nil {
		t.Errorf("NoopMailer.SendLicense returned error: %v", err)
	}
}

func TestNewResendMailer_RejectsEmptyKey(t *testing.T) {
	if _, err := NewResendMailer("", "from@example.com", "", ""); err == nil {
		t.Error("expected error for empty api key")
	}
}

func TestNewResendMailer_RejectsEmptyFrom(t *testing.T) {
	if _, err := NewResendMailer("re_test", "", "", ""); err == nil {
		t.Error("expected error for empty from")
	}
}

func TestNewResendMailer_DefaultsSupportEmailToFromAddress(t *testing.T) {
	m, err := NewResendMailer("re_test", "Aura Optimizer <licenses@mail.example.com>", "", "")
	if err != nil {
		t.Fatalf("constructor: %v", err)
	}
	if m.supportEmail != "licenses@mail.example.com" {
		t.Errorf("supportEmail = %q, want extracted from From", m.supportEmail)
	}
}

func TestResendMailer_SendLicense_RejectsEmptyRecipient(t *testing.T) {
	m, err := NewResendMailer("re_test", "from@example.com", "support@example.com", "https://example.com/download")
	if err != nil {
		t.Fatalf("constructor: %v", err)
	}
	if err := m.SendLicense(context.Background(), "", "key", "product"); err == nil {
		t.Error("expected error for empty recipient")
	}
}

func TestRenderLicenseHTML_SubstitutesAllVariables(t *testing.T) {
	tmpl := `<a href="{{DOWNLOAD_URL}}">go</a> key={{LICENSE_KEY}} support={{SUPPORT_EMAIL}} to={{CUSTOMER_EMAIL}}`
	got := renderLicenseHTML(tmpl, map[string]string{
		"LICENSE_KEY":    "AURA-XXXXX-YYYYY-ZZZZZ",
		"SUPPORT_EMAIL":  "help@example.com",
		"CUSTOMER_EMAIL": "buyer@example.com",
		"DOWNLOAD_URL":   "https://example.com/download.exe",
	})
	for _, want := range []string{
		"AURA-XXXXX-YYYYY-ZZZZZ",
		"help@example.com",
		"buyer@example.com",
		"https://example.com/download.exe",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered output missing %q\n--- got ---\n%s", want, got)
		}
	}
	if strings.Contains(got, "{{") {
		t.Errorf("output still contains unsubstituted {{...}}: %s", got)
	}
}

func TestRenderLicenseHTML_EscapesHTMLInValues(t *testing.T) {
	tmpl := `customer: {{CUSTOMER_EMAIL}}`
	got := renderLicenseHTML(tmpl, map[string]string{
		"CUSTOMER_EMAIL": `attacker@evil.com" onerror="alert(1)`,
	})
	if strings.Contains(got, `"`) || strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("HTML special chars not escaped: %s", got)
	}
}

func TestAddressPart(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Aura Optimizer <licenses@mail.example.com>", "licenses@mail.example.com"},
		{"bare@example.com", "bare@example.com"},
		{"   trimmed@example.com   ", "trimmed@example.com"},
		{"Some Name <x@y.com>", "x@y.com"},
	}
	for _, tc := range tests {
		if got := addressPart(tc.in); got != tc.want {
			t.Errorf("addressPart(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLicenseEmailTemplate_ContainsExpectedVariables(t *testing.T) {
	tmpl := LicenseEmailTemplate()
	for _, v := range []string{"{{LICENSE_KEY}}", "{{SUPPORT_EMAIL}}", "{{CUSTOMER_EMAIL}}", "{{DOWNLOAD_URL}}"} {
		if !strings.Contains(tmpl, v) {
			t.Errorf("embedded template missing %s", v)
		}
	}
}
