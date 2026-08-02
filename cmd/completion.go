package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/store"
	"github.com/spf13/cobra"
)

const cacheFileName = ".services_cache"

var installShell string

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  source <(upass completion bash)

Zsh:
  source <(upass completion zsh)

Fish:
  upass completion fish | source`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completion permanently",
	Long: `Detect current shell and install completion script to the correct location.
	After installation, completion will work in new terminal sessions.
	Supported shells: bash, zsh, fish`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := installBinary(); err != nil {
			return fmt.Errorf("install binary: %w", err)
		}

		var shell string
		if installShell != "" {
			shell = installShell
			fmt.Printf("Using forced shell: %s\n", shell)
		} else {
			shell = detectShell()
			if shell == "" {
				common.YellowPrintln("⚠️ Could not detect shell. Please use --shell flag:")
				fmt.Println(" upass completion install --shell fish")
				return nil
			}
			fmt.Printf("Detected shell: %s\n", shell)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}

		switch shell {
		case "bash":
			dir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create bash completion directory: %w", err)
			}
			path := filepath.Join(dir, "upass")
			if err := rootCmd.GenBashCompletionFile(path); err != nil {
				return fmt.Errorf("failed to generate bash completion file: %w", err)
			}
			common.GreenPrintf("Bash completion installed to %s\n", path)

		case "zsh":
			dir := filepath.Join(home, ".zfunc")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create zsh completion directory: %w", err)
			}
			path := filepath.Join(dir, "_upass")
			if err := rootCmd.GenZshCompletionFile(path); err != nil {
				return fmt.Errorf("failed to generate zsh completion file: %w", err)
			}
		case "fish":
			dir := filepath.Join(home, ".config", "fish", "completions")
			os.MkdirAll(dir, 0755)
			path := filepath.Join(dir, "upass.fish")
			rootCmd.GenFishCompletionFile(path, true)
			common.GreenPrintf("Fish completion installed to %s\n", path)

		default:
			common.YellowPrintln("Unknown shell. Completion not installed.")
			fmt.Println("Use 'upass completion [bash|zsh|fish]' for manual setup.")
		}

		if err := ensureShellConfig(shell, home); err != nil {
			common.YellowPrintf("Could not auto-configure shell: %v\n", err)
			fmt.Println("   Please configure your shell manually as shown above.")
		}

		common.GreenPrintf("upass installed to %s\n", filepath.Join(home, ".local", "bin", "upass"))
		fmt.Println()
		fmt.Println("Open a new terminal and run: upass init")
		return nil
	},
}

func serviceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var services []string

	if vaultService != nil {
		services = vaultService.ListServices()
	}

	if services == nil {
		services = loadServicesCache()
	}

	if len(services) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	toComplete = strings.ToLower(toComplete)
	var matches []string
	for _, s := range services {
		if strings.HasPrefix(strings.ToLower(s), toComplete) {
			matches = append(matches, s)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

// loadServicesCache reads the cached list of services used for completion.
// Returns nil if the cache file is missing or empty.
func loadServicesCache() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	cachePath := filepath.Join(home, store.DefaultVaultDir, cacheFileName)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func detectShell() string {
	if os.Getenv("FISH_VERSION") != "" || os.Getenv("__fish_bin_dir") != "" {
		return "fish"
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		ppid := os.Getppid()
		if runtime.GOOS == "linux" {
			commPath := filepath.Join("/proc", fmt.Sprintf("%d", ppid), "comm")
			if commBytes, err := os.ReadFile(commPath); err == nil {
				comm := strings.TrimSpace(string(commBytes))
				if strings.Contains(comm, "fish") {
					return "fish"
				}
				if strings.Contains(comm, "zsh") {
					return "zsh"
				}
				if strings.Contains(comm, "bash") {
					return "bash"
				}
			}
		}
	}

	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}

	return ""
}

// installBinary copies the current executable to a user-local bin directory.
// It avoids using sudo by default, making it safe for CI/CD, containers, and
// standard users.
func installBinary() error {
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find current executable: %w", err)
	}

	var targetDir string
	var targetName string
	if runtime.GOOS == "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}
		targetDir = filepath.Join(home, "bin")
		targetName = "upass.exe"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}
		targetDir = filepath.Join(home, ".local", "bin")
		targetName = "upass"
	}

	targetPath := filepath.Join(targetDir, targetName)

	if currentPath == targetPath {
		return nil
	}

	if _, err := os.Stat(targetPath); err == nil {
		common.YellowPrintf("Warning: %s already exists and will be overwritten.\n", targetPath)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	data, err := os.ReadFile(currentPath)
	if err != nil {
		return fmt.Errorf("failed to read current executable: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0755); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s.\nTry running with sudo, or copy manually:\n  sudo cp %s %s", targetPath, currentPath, targetPath)
		}
		return fmt.Errorf("failed to write executable to %s: %w", targetPath, err)
	}

	return nil
}

func init() {
	completionCmd.AddCommand(completionInstallCmd)
	rootCmd.AddCommand(completionCmd)
	completionInstallCmd.Flags().StringVar(&installShell, "shell", "",
		"Force shell type (bash, zsh, fish). If not set, auto-detect.")
}

// ensureShellConfig checks if the required PATH and completion lines exist
// in the user's shell config file. If not, it safely appends them.
func ensureShellConfig(shell string, homeDir string) error {
	if runtime.GOOS == "windows" {
		return ensureWindowsConfig(homeDir)
	}

	var configPath string
	var requiredLines []string

	binDir := filepath.Join(homeDir, ".local", "bin")

	switch shell {
	case "zsh":
		configPath = filepath.Join(homeDir, ".zshrc")
		requiredLines = []string{
			fmt.Sprintf(`export PATH="%s:$PATH"`, binDir),
			"fpath=(~/.zfunc $fpath)",
			"autoload -Uz compinit && compinit",
		}
	case "bash":
		configPath = filepath.Join(homeDir, ".bashrc")
		requiredLines = []string{
			fmt.Sprintf(`export PATH="%s:$PATH"`, binDir),
		}
	case "fish":
		configPath = filepath.Join(homeDir, ".config", "fish", "config.fish")
		requiredLines = []string{
			fmt.Sprintf(`set -gx PATH %s $PATH`, binDir),
		}
	default:
		return nil
	}

	if configPath == "" || len(requiredLines) == 0 {
		return nil
	}

	file, err := os.Open(configPath)
	var existingContent string
	if err != nil {
		if os.IsNotExist(err) {
			common.YellowPrintf("Creating %s to configure PATH...\n", configPath)
			if shell == "fish" {
				os.MkdirAll(filepath.Dir(configPath), 0755)
			}
			return appendLinesToFile(configPath, requiredLines)
		}
		return fmt.Errorf("failed to open %s: %w", configPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		existingContent += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	var linesToAdd []string
	for _, req := range requiredLines {
		if !strings.Contains(existingContent, req) {
			linesToAdd = append(linesToAdd, req)
		}
	}

	if len(linesToAdd) > 0 {
		if len(linesToAdd) > 0 {
			common.YellowPrintln("⚠️  UPass needs to modify your shell configuration to add PATH and auto-completion.")
			common.YellowPrintf("   The following file will be updated: %s\n", configPath)
			common.YellowPrintln("   If you ever need to undo this, simply open the file and remove the lines added by UPass.")
			fmt.Print("   Do you want to proceed? (y/n): ")
		}

		if !readConfirmation() {
			common.YellowPrintln("Skipped. You will need to configure your shell manually.")
			return nil
		}

		common.YellowPrintf("Automatically configuring %s for PATH and completion...\n", shell)
		if err := appendLinesToFile(configPath, linesToAdd); err != nil {
			return err
		}
		common.GreenPrintf("Configuration added to %s\n", configPath)
		common.YellowPrintf("Please restart your terminal or run: source %s", configPath)
	}

	return nil
}

// ensureWindowsConfig handles PATH configuration for Windows.
// On Windows, we cannot easily modify system PATH programmatically without admin rights.
// Instead, we provide clear instructions to the user.
func ensureWindowsConfig(homeDir string) error {
	binDir := filepath.Join(homeDir, "bin")

	pathEnv := os.Getenv("PATH")
	if strings.Contains(pathEnv, binDir) {
		return nil
	}

	fmt.Println()
	common.YellowPrintln("IMPORTANT: Add the following directory to your system PATH:")
	fmt.Printf("   %s\n", binDir)
	fmt.Println()
	fmt.Println("   To do this:")
	fmt.Println("   1. Press Win+R, type 'sysdm.cpl', press Enter")
	fmt.Println("   2. Go to 'Advanced' tab → 'Environment Variables'")
	fmt.Println("   3. Under 'User variables', select 'Path' → 'Edit'")
	fmt.Println("   4. Click 'New' and paste the path above")
	fmt.Println("   5. Click OK, then restart your terminal")

	return nil
}

// appendLinesToFile safely appends lines to a file, adding a newline if the file is
// not empty.
func appendLinesToFile(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err == nil && info.Size() > 0 {
		f.WriteString("\n")
	}

	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}
