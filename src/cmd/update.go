package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Public LibreSeal CLI locations. Nothing is downloaded from these URLs by the
// CLI itself: they are only printed or opened in the user's browser.
const (
	RepoURL = "https://github.com/Dos2Locos/libreseal-cli"
	DocsURL = RepoURL + "#readme"
)

func init() {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "🆙 Show how to update the LibreSeal CLI",
		Long: "Prints the steps to update the LibreSeal CLI from source. " +
			"The CLI never downloads or executes installer scripts by itself.",
		RunE: func(cmd *cobra.Command, args []string) error {
			printUpdateInstructions(cmd.OutOrStdout())
			return nil
		},
	}
	rootCmd.AddCommand(updateCmd)
}

func printUpdateInstructions(w io.Writer) {
	fmt.Fprintf(w, `To update the LibreSeal CLI, rebuild it from a release tag of
%s:

  git clone %s.git   # or: git -C libreseal-cli fetch --tags
  cd libreseal-cli
  git checkout <tag>                 # e.g. the latest tag from 'git tag --sort=-v:refname'
  ./scripts/install-from-source.sh

Release notes: %s/releases
`, RepoURL, RepoURL, RepoURL)
}
