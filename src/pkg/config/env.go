package config

import (
	"os"
	"strings"

	"github.com/phasehq/golang-sdk/v2/phase/misc"
)

// Env returns the value of LIBRESEAL_<name>, falling back to PHASE_<name> for
// compatibility with the upstream Phase CLI. LIBRESEAL_* takes precedence.
// Example: Env("HOST") reads LIBRESEAL_HOST, then PHASE_HOST.
func Env(name string) string {
	if v := os.Getenv("LIBRESEAL_" + name); v != "" {
		return v
	}
	return os.Getenv("PHASE_" + name)
}

// ConfigureSSLVerification reads LIBRESEAL_VERIFY_SSL (or PHASE_VERIFY_SSL) and
// disables TLS certificate verification in the SDK when it is set to "false"
// (case-insensitive). Any other value — or the variable being unset — keeps
// verification enabled. This mirrors the Python CLI's behavior
// (os.environ.get("PHASE_VERIFY_SSL", "True").lower() != "false") and makes
// the hint shown in SSL error messages ("You may set PHASE_VERIFY_SSL=False
// to bypass this check") actually work.
//
// It must run before any SDK network call: the SDK caches its HTTP client on
// first use, so changes to misc.VerifySSL after that have no effect.
func ConfigureSSLVerification() {
	if strings.EqualFold(Env("VERIFY_SSL"), "false") {
		misc.VerifySSL = false
	}
}
