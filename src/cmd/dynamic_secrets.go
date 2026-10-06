package cmd

import (
	"errors"
	"net/http"

	"github.com/phasehq/golang-sdk/v2/phase/network"
	"github.com/spf13/cobra"
)

var dynamicSecretsCmd = &cobra.Command{
	Use:   "dynamic-secrets",
	Short: "⚡️ Manage dynamic secrets",
}

var dynamicSecretsLeaseCmd = &cobra.Command{
	Use:   "lease",
	Short: "📜 Manage dynamic secret leases",
}

func init() {
	dynamicSecretsCmd.AddCommand(dynamicSecretsLeaseCmd)
	rootCmd.AddCommand(dynamicSecretsCmd)
}

// errDynamicSecretsUnavailable is returned when the server does not expose the
// dynamic secrets API. LibreSeal servers do not implement it (the upstream
// implementation is under a proprietary license).
var errDynamicSecretsUnavailable = errors.New(
	"dynamic secrets are not available on this server (LibreSeal does not implement them)")

// wrapDynamicSecretsErrors turns "not found" responses from every
// dynamic-secrets subcommand into errDynamicSecretsUnavailable.
func wrapDynamicSecretsErrors(parent *cobra.Command) {
	for _, sub := range parent.Commands() {
		wrapDynamicSecretsErrors(sub)
		if sub.RunE == nil {
			continue
		}
		run := sub.RunE
		sub.RunE = func(cmd *cobra.Command, args []string) error {
			return translateDynamicSecretsError(run(cmd, args))
		}
	}
}

func translateDynamicSecretsError(err error) error {
	var apiErr *network.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return errDynamicSecretsUnavailable
	}
	return err
}
