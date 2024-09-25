package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "git-ai",
	Short: "A git commit helper that uses ChatGPT to generate commit messages and PR descriptions",
}

var (
	commitCmd        *cobra.Command
	configCmd        *cobra.Command
	prDescriptionCmd *cobra.Command
	releaseCmd       *cobra.Command
)

func init() {
	commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Generate a commit message and create a commit",
		Run:   runCommit,
	}

	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Configure OpenAI API credentials",
		Run:   runConfig,
	}

	prDescriptionCmd = &cobra.Command{
		Use:   "pr [base-branch]",
		Short: "Generate a structured PR description based on differences between the current branch and the specified base branch",
		Args:  cobra.MaximumNArgs(1),
		Run:   runPRDescription,
	}

	releaseCmd = &cobra.Command{
		Use:   "release [previous-version] [new-version]",
		Short: "Generate a structured release description based on differences between two specified versions",
		Args:  cobra.ExactArgs(2),
		Run:   runReleaseDescription,
	}

	rootCmd.AddCommand(commitCmd, configCmd, prDescriptionCmd)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME/.project-commit")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runConfig(cmd *cobra.Command, args []string) {
	var apiKey string
	fmt.Print("Enter your OpenAI API key: ")
	fmt.Scanln(&apiKey)

	viper.Set("openai_api_key", apiKey)

	configPath := filepath.Join(os.Getenv("HOME"), ".project-commit")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}

	if err := viper.WriteConfigAs(filepath.Join(configPath, "config.json")); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Println("Configuration saved successfully.")
}

func runCommit(cmd *cobra.Command, args []string) {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No config file found. Please run 'project-commit config' to set up your API key.")
		return
	}

	apiKey := viper.GetString("openai_api_key")
	if apiKey == "" {
		fmt.Println("API key not found. Please run 'project-commit config' to set up your API key.")
		return
	}

	changes, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		fmt.Println("Error executing git status:", err)
		return
	}

	if len(strings.TrimSpace(string(changes))) == 0 {
		fmt.Println("No changes detected. Nothing to commit.")
		return
	}

	diff, err := exec.Command("git", "diff").Output()
	if err != nil {
		fmt.Println("Error executing git diff:", err)
		return
	}

	if string(diff) != "" {
		stagedFiles, err := exec.Command("git", "diff", "--name-only", "--cached").Output()
		if err != nil {
			fmt.Println("Error getting staged files:", err)
			return
		}

		if len(strings.TrimSpace(string(stagedFiles))) > 0 {
			fmt.Println("\nStaged files:")
			fmt.Println(string(stagedFiles))
		}

		title := strings.Split(string(diff), "\n")[0]

		description, err := generateCommitMessage(string(diff), title, apiKey)
		if err != nil {
			fmt.Println("Error generating commit message:", err)
			return
		}

		fmt.Println("\nGenerated commit message:")
		fmt.Println(description)

		fmt.Print("\nDo you want to proceed with git add . and commit? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled.")
			return
		}

		cmd := exec.Command("git", "add", ".")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error executing git add .:", err)
			return
		}

		cmd = exec.Command("git", "commit", "-m", description)
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error executing git commit:", err)
			return
		}

		fmt.Println("Changes added and committed successfully!")

	} else {
		fmt.Println("No changes detected. Nothing to commit.")
		return
	}
}

func runPRDescription(cmd *cobra.Command, args []string) {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No config file found. Please run 'project-commit config' to set up your API key.")
		return
	}

	apiKey := viper.GetString("openai_api_key")
	if apiKey == "" {
		fmt.Println("API key not found. Please run 'project-commit config' to set up your API key.")
		return
	}

	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Println("Error getting current branch:", err)
		return
	}

	var baseBranch string
	if len(args) > 0 {
		baseBranch = args[0]
	} else {
		baseBranch = getDefaultBaseBranch()
	}

	diff, err := exec.Command("git", "diff", baseBranch+".."+currentBranch).Output()
	if err != nil {
		fmt.Println("Error executing git diff:", err)
		return
	}

	if string(diff) == "" {
		fmt.Println("No differences detected between the current branch and", baseBranch)
		return
	}

	description, err := generatePRDescription(string(diff), baseBranch, currentBranch, apiKey)
	if err != nil {
		fmt.Println("Error generating PR description:", err)
		return
	}

	fmt.Println("\nGenerated PR Description:")
	fmt.Println(description)
}

func runReleaseDescription(cmd *cobra.Command, args []string) {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No config file found. Please run 'project-commit config' to set up your API key.")
		return
	}

	apiKey := viper.GetString("openai_api_key")
	if apiKey == "" {
		fmt.Println("API key not found. Please run 'project-commit config' to set up your API key.")
		return
	}

	previousVersion := args[0]
	newVersion := args[1]

	diff, err := exec.Command("git", "diff", previousVersion+".."+newVersion).Output()
	if err != nil {
		fmt.Println("Error executing git diff:", err)
		return
	}

	if string(diff) == "" {
		fmt.Println("No differences detected between", previousVersion, "and", newVersion)
		return
	}

	description, err := generateReleaseDescription(string(diff), previousVersion, newVersion, apiKey)
	if err != nil {
		fmt.Println("Error generating release description:", err)
		return
	}

	fmt.Println("\nGenerated Release Description:")
	fmt.Println(description)
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getDefaultBaseBranch() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "main")
	if err := cmd.Run(); err == nil {
		return "main"
	}

	cmd = exec.Command("git", "rev-parse", "--verify", "master")
	if err := cmd.Run(); err == nil {
		return "master"
	}

	return ""
}

func generateCommitMessage(diff, title, apiKey string) (string, error) {
	prompt := fmt.Sprintf("Based on the following git diff and the context '%s', generate a concise commit message in English. Start with a brief summary line (50 characters or less), followed by a blank line, and then a more detailed explanation. Do not include any prefixes like 'title:', '**Summary:**','**Details:**:'. Here's the diff:\n\n%s", title, diff)
	return callOpenAI(prompt, apiKey)
}

func generatePRDescription(diff, baseBranch, compareBranch, apiKey string) (string, error) {
	prompt := fmt.Sprintf("Generate a structured GitHub Pull Request description based on the following git diff between '%s' and '%s' branches. Use this format:\n\n## PR Title\n\n## PR Description\n\n## Changes Made\n\n## Instructions for Reviewer\n\n## Recommendations for Testing this PR\n\n## Concerns\n\n## Link to Related Issue(s)\n\nEnsure the content is clear, concise, and relevant to each heading. Here's the diff:\n\n%s", baseBranch, compareBranch, diff)
	return callOpenAI(prompt, apiKey)
}

func generateReleaseDescription(diff, previousVersion, newVersion, apiKey string) (string, error) {
	currentDate := time.Now().Format("2006.01.02")
	releaseName := fmt.Sprintf("Release v%s", currentDate)

	prompt := fmt.Sprintf(`Generate a structured release description based on the following git diff between versions '%s' and '%s'. Use this exact format:

%s
## 🚀 What's New
- [New feature 1]
- [New feature 2]
- [New feature 3]
## 🐛 Bug Fixes
- [Bug fix 1]
- [Bug fix 2]
- [Bug fix 3]
## 🛠 Changes
- [Change 1]
- [Change 2]
- [Change 3]
## 📚 Documentation
- [Documentation update 1]
- [Documentation update 2]
## 🏗 Dependencies
- [Dependency update 1]
- [Dependency update 2]
## 🗑 Deprecations
- [Deprecated item 1]
- [Deprecated item 2]
---
**Installation**: [Installation or update instructions]
**Additional Notes**: [Any additional relevant information]
**Contributors**: [@user1], [@user2], [@user3]

Ensure the content is clear, concise, and relevant to each heading. Fill in the placeholders with actual content based on the diff. If there's no relevant information for a section, include "No updates in this release." Here's the diff:

%s`, previousVersion, newVersion, releaseName, diff)
	return callOpenAI(prompt, apiKey)
}

func callOpenAI(prompt, apiKey string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a specialized assistant for generating git commit messages, PR descriptions, and release notes. Focus on generating content in English, formatted in Markdown when applicable. Emphasize clarity, conciseness, and relevance in your responses.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	message := result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return message, nil
}
