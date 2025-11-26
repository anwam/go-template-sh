package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/anwam/go-template-sh/internal/generator"
	"github.com/anwam/go-template-sh/internal/prompt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-template-sh",
	Short: "Generate Go HTTP server project templates",
	Long: `A CLI tool to scaffold production-ready Go HTTP server projects.
Follows twelve-factor app methodology and Go community best practices.
Includes complete observability (logging, tracing, metrics).`,
	RunE: runGenerate,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP("output", "o", ".", "Output directory for the generated project")
	rootCmd.Flags().StringP("name", "n", "", "Project name (if not provided, will prompt)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	outputDir, _ := cmd.Flags().GetString("output")
	projectName, _ := cmd.Flags().GetString("name")

	fmt.Println("🚀 Welcome to go-template-sh - Go HTTP Server Template Generator")
	fmt.Println()

	config, err := prompt.CollectConfiguration(projectName)
	if err != nil {
		return fmt.Errorf("failed to collect configuration: %w", err)
	}

	confirmed := false
	prompt := &survey.Confirm{
		Message: "Ready to generate your project. Continue?",
		Default: true,
	}
	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Project generation cancelled")
		return nil
	}

	gen := generator.New(config, outputDir)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Project generated successfully!")
	fmt.Printf("📁 Location: %s/%s\n", outputDir, config.ProjectName)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", config.ProjectName)
	fmt.Println("  go mod download")
	fmt.Println("  make run")
	fmt.Println()

	return nil
}
