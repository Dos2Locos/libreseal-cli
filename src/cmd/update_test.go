package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpdatePrintsSourceInstructionsWithoutPhaseInfrastructure(t *testing.T) {
	var out bytes.Buffer
	printUpdateInstructions(&out)
	got := out.String()

	if !strings.Contains(got, "https://github.com/Dos2Locos/libreseal-cli") {
		t.Fatalf("update instructions should point to the LibreSeal CLI repo:\n%s", got)
	}
	if !strings.Contains(got, "install-from-source.sh") {
		t.Fatalf("update instructions should use the source installer:\n%s", got)
	}
	if strings.Contains(got, "phase.dev") {
		t.Fatalf("update instructions must not reference Phase infrastructure:\n%s", got)
	}
}
