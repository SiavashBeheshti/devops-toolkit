package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// newCompletionCmd creates the completion command
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for devops-toolkit.

Supported shells:
  bash        Bash completion script
  zsh         Zsh completion script
  fish        Fish completion script
  powershell  PowerShell completion script

To load completions:

Bash:
  # Linux:
  $ devops-toolkit completion bash > /etc/bash_completion.d/devops-toolkit

  # macOS:
  $ devops-toolkit completion bash > $(brew --prefix)/etc/bash_completion.d/devops-toolkit

  # Or load for current session only:
  $ source <(devops-toolkit completion bash)

Zsh:
  # If shell completion is not already enabled in your environment:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Generate and load completions:
  $ devops-toolkit completion zsh > "${fpath[1]}/_devops-toolkit"

  # Or for Oh My Zsh:
  $ devops-toolkit completion zsh > ~/.oh-my-zsh/completions/_devops-toolkit

  # Or load for current session only:
  $ source <(devops-toolkit completion zsh)

Fish:
  $ devops-toolkit completion fish > ~/.config/fish/completions/devops-toolkit.fish

  # Or load for current session only:
  $ devops-toolkit completion fish | source

PowerShell:
  PS> devops-toolkit completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, add the output to your profile:
  PS> devops-toolkit completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}

