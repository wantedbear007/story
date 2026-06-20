package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/infrastructure/config"
	"github.com/anomalyco/story/internal/infrastructure/llm"
)

func newLLMCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "llm-check",
		Short: "Test LLM provider connectivity and response",
		Long:  "Verify the LLM provider is configured and send a test prompt to check response quality.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLLMCheck()
		},
	}
}

func runLLMCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfgPath := resolveConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if !hasLLMConfig(cfg) {
		fmt.Println("✗ LLM is not configured")
		fmt.Println("  Run 'story llm-config' to set up an LLM provider.")
		return nil
	}

	provider, err := llm.NewProvider(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating LLM provider: %w", err)
	}
	adapter := llm.NewCompleteAdapter(provider)

	fmt.Printf("LLM Provider: %s (%s)\n", provider.Name(), cfg.LLM.Provider)
	fmt.Println("Sending test prompt...")

	start := time.Now()
	response, err := adapter.Complete(ctx, "Say 'Hello from Story!' in one short sentence.", 100)
	elapsed := time.Since(start)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	fmt.Printf("Response (%v):\n", elapsed.Round(time.Millisecond))
	fmt.Println(response)
	return nil
}
