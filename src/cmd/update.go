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
		Long: "Prints the steps to update the LibreSeal CLI from a verified release binary " +
			"or from source. The CLI never downloads or executes anything by itself.",
		RunE: func(cmd *cobra.Command, args []string) error {
			printUpdateInstructions(cmd.OutOrStdout())
			return nil
		},
	}
	rootCmd.AddCommand(updateCmd)
}

func printUpdateInstructions(w io.Writer) {
	fmt.Fprintf(w, `To update the LibreSeal CLI, either:

1. Download the binary for your platform from %s/releases,
   verify checksums.txt with cosign and sha256sum (see %s), and replace
   this executable with it; or

2. Rebuild it from a release tag:

  git clone %s.git   # or: git -C libreseal-cli fetch --tags
  cd libreseal-cli
  git checkout <tag>                 # e.g. the latest tag from 'git tag --sort=-v:refname'
  ./scripts/install-from-source.sh
`, RepoURL, DocsURL, RepoURL)
}
