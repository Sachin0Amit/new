package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ANSI Color Codes for Rich Styling
const (
	ColorCyan   = "\033[36m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorGreen  = "\033[32m"
	ColorGray   = "\033[90m"
	ColorReset  = "\033[0m"
	Bold        = "\033[1m"
)

// SovereignShell provides an interactive REPL for direct cognitive interaction.
type SovereignShell struct {
	UserPrompt string
	CoreName   string
}

// NewShell initializes a shell with custom branding.
func NewShell() *SovereignShell {
	return &SovereignShell{
		UserPrompt: Bold + ColorCyan + "sovereign> " + ColorReset,
		CoreName:   Bold + ColorPurple + "TITAN-V1" + ColorReset,
	}
}

// Run initiates the interactive REPL loop.
func (s *SovereignShell) Run() {
	s.printBanner()
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(s.UserPrompt)
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println(ColorGray + "Disconnecting from core..." + ColorReset)
			break
		}

		s.processInput(input)
	}
}

func (s *SovereignShell) printBanner() {
	fmt.Println(Bold + ColorBlue + "╔══════════════════════════════════════════════════╗")
	fmt.Println("║          SOVEREIGN INTELLIGENCE SHELL            ║")
	fmt.Println("╚══════════════════════════════════════════════════╝" + ColorReset)
	fmt.Printf("%s Local core connected at %s\n", ColorGray, time.Now().Format(time.Kitchen))
	fmt.Println("Type 'exit' to disconnect.")
}

func (s *SovereignShell) processInput(input string) {
	fmt.Printf("%sThinking...%s\n", ColorGray, ColorReset)
	
	// Simulation of derivation (in production this calls the REST API/Orchestrator)
	time.Sleep(500 * time.Millisecond)
	
	fmt.Printf("[%s] %sThe probability space suggests: %s%s\n\n", 
		s.CoreName, ColorGreen, ColorReset, "Action confirmed. Sovereignty maintained.")
}
