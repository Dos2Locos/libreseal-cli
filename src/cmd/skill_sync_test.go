package cmd

import (
	"regexp"
	"strings"
	"testing"

	"github.com/phasehq/cli/pkg/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Every `libreseal ...` invocation in the embedded AI skill must name a real
// command and only use flags that command accepts, so agents are never told
// to run something the CLI does not support.
func TestSkillCommandsExistInCLI(t *testing.T) {
	content := ai.SkillContent()
	invocation := regexp.MustCompile("libreseal((?: [a-z][a-z-]*)+)((?: [^`|\\n]*)?)")
	flagRe := regexp.MustCompile(`--([a-z][a-z-]*)`)

	matches := invocation.FindAllStringSubmatch(content, -1)
	if len(matches) < 10 {
		t.Fatalf("expected the skill to document many commands, found %d", len(matches))
	}

	for _, m := range matches {
		words := strings.Fields(m[1])
		cmd, rest, err := rootCmd.Find(words)
		if err != nil || cmd == rootCmd {
			t.Errorf("skill references unknown command: libreseal %s", m[1])
			continue
		}
		_ = rest
		for _, f := range flagRe.FindAllStringSubmatch(m[2], -1) {
			if lookupFlag(cmd, f[1]) == nil {
				t.Errorf("skill uses unknown flag --%s for 'libreseal %s'", f[1], cmd.CommandPath())
			}
		}
	}
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if f := cmd.Flags().Lookup(name); f != nil {
		return f
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil {
		return f
	}
	return nil
}

func TestSkillMentionsLeastPrivilegeAndNoPhaseCloud(t *testing.T) {
	content := ai.SkillContent()
	for _, want := range []string{"service account", "LIBRESEAL_SERVICE_TOKEN", "defence in depth"} {
		if !strings.Contains(content, want) {
			t.Errorf("skill should mention %q", want)
		}
	}
	for _, banned := range []string{"phase.dev", "Phase Cloud", "phase run", "phase secrets"} {
		if strings.Contains(content, banned) {
			t.Errorf("skill must not reference %q", banned)
		}
	}
}
