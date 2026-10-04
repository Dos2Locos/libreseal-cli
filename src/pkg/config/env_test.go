package config

import (
	"os"
	"testing"

	"github.com/phasehq/golang-sdk/v2/phase/misc"
)

func TestConfigureSSLVerification(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		unset      bool
		wantVerify bool
	}{
		{name: "unset keeps verification enabled", unset: true, wantVerify: true},
		{name: "empty keeps verification enabled", value: "", wantVerify: true},
		{name: "False disables verification", value: "False", wantVerify: false},
		{name: "false disables verification", value: "false", wantVerify: false},
		{name: "FALSE disables verification", value: "FALSE", wantVerify: false},
		{name: "true keeps verification enabled", value: "true", wantVerify: true},
		{name: "True keeps verification enabled", value: "True", wantVerify: true},
		{name: "zero keeps verification enabled", value: "0", wantVerify: true},
		{name: "no keeps verification enabled", value: "no", wantVerify: true},
		{name: "arbitrary value keeps verification enabled", value: "banana", wantVerify: true},
	}

	original := misc.VerifySSL
	t.Cleanup(func() { misc.VerifySSL = original })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			misc.VerifySSL = true // reset the global before each case

			// t.Setenv registers restoration of the original value; for the
			// unset case we additionally clear it after registration.
			t.Setenv("PHASE_VERIFY_SSL", tt.value)
			if tt.unset {
				os.Unsetenv("PHASE_VERIFY_SSL")
			}

			ConfigureSSLVerification()

			if misc.VerifySSL != tt.wantVerify {
				t.Errorf("PHASE_VERIFY_SSL=%q: misc.VerifySSL = %v, want %v", tt.value, misc.VerifySSL, tt.wantVerify)
			}
		})
	}
}

func TestEnvPrefersLibresealPrefix(t *testing.T) {
	t.Setenv("LIBRESEAL_SERVICE_TOKEN", "ls-token")
	t.Setenv("PHASE_SERVICE_TOKEN", "phase-token")
	if got := Env("SERVICE_TOKEN"); got != "ls-token" {
		t.Fatalf("Env(SERVICE_TOKEN) = %q, want LIBRESEAL_ value", got)
	}
}

func TestEnvFallsBackToPhasePrefix(t *testing.T) {
	t.Setenv("LIBRESEAL_HOST", "")
	t.Setenv("PHASE_HOST", "https://legacy.example.com")
	if got := Env("HOST"); got != "https://legacy.example.com" {
		t.Fatalf("Env(HOST) = %q, want PHASE_ fallback", got)
	}
}

func TestConfigureSSLVerificationHonoursLibresealVariable(t *testing.T) {
	t.Setenv("PHASE_VERIFY_SSL", "")
	t.Setenv("LIBRESEAL_VERIFY_SSL", "False")
	orig := misc.VerifySSL
	t.Cleanup(func() { misc.VerifySSL = orig })
	misc.VerifySSL = true
	ConfigureSSLVerification()
	if misc.VerifySSL {
		t.Fatal("LIBRESEAL_VERIFY_SSL=False should disable verification")
	}
}
