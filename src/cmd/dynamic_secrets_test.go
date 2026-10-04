package cmd

import (
	"errors"
	"testing"

	"github.com/phasehq/golang-sdk/v2/phase/network"
)

func TestDynamicSecretsNotFoundIsReportedAsUnavailable(t *testing.T) {
	err := translateDynamicSecretsError(&network.APIError{StatusCode: 404})
	if !errors.Is(err, errDynamicSecretsUnavailable) {
		t.Fatalf("got %v, want errDynamicSecretsUnavailable", err)
	}
}

func TestDynamicSecretsOtherErrorsPassThrough(t *testing.T) {
	orig := &network.APIError{StatusCode: 500}
	if err := translateDynamicSecretsError(orig); err != orig {
		t.Fatalf("got %v, want the original error", err)
	}
	if translateDynamicSecretsError(nil) != nil {
		t.Fatal("nil must stay nil")
	}
}
