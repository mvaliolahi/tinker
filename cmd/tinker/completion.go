package main

import (
        "fmt"
        "os"

        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

func completionCmd(root *cobra.Command) *cobra.Command {
        cmd := &cobra.Command{
                Use:   "completion [bash|zsh|fish|powershell]",
                Short: "Generate shell completion script",
                Long: `Generate shell completion script for tinker.

Load completions:
  Bash:   echo 'source <(tinker completion bash)' >> ~/.bashrc
  Zsh:    echo 'source <(tinker completion zsh)' >> ~/.zshrc
  Fish:   tinker completion fish > ~/.config/fish/completions/tinker.fish
  PS:     tinker completion powershell | Out-String | Invoke-Expression`,
                ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
                Args:      cobra.ExactArgs(1),
                Run: func(_ *cobra.Command, args []string) {
                        switch args[0] {
                        case "bash":
                                _ = root.GenBashCompletion(os.Stdout)
                        case "zsh":
                                _ = root.GenZshCompletion(os.Stdout)
                        case "fish":
                                _ = root.GenFishCompletion(os.Stdout, true)
                        case "powershell":
                                _ = root.GenPowerShellCompletionWithDesc(os.Stdout)
                        default:
                                fmt.Println(ui.Error("unsupported shell: " + args[0]))
                        }
                },
        }
        return cmd
}
