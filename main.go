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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "project-commit",
	Short: "A git commit helper that uses ChatGPT to generate commit messages",
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate a commit message and create a commit",
	Run:   runCommit,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure OpenAI API credentials",
	Run:   runConfig,
}

func init() {
	rootCmd.AddCommand(commitCmd, configCmd)
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

	// Get changes (both staged and unstaged)
	changes, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		fmt.Println("Error executing git status:", err)
		return
	}

	// Check if there are any changes
	if len(strings.TrimSpace(string(changes))) == 0 {
		fmt.Println("No changes detected. Nothing to commit.")
		return
	}

	// Get detailed diff for generating commit message
	diff, err := exec.Command("git", "diff").Output()
	if err != nil {
		fmt.Println("Error executing git diff:", err)
		return
	}

	if string(diff) != "" {
		// Get staged files
		stagedFiles, err := exec.Command("git", "diff", "--name-only", "--cached").Output()
		if err != nil {
			fmt.Println("Error getting staged files:", err)
			return
		}

		// If there are staged files, show them
		if len(strings.TrimSpace(string(stagedFiles))) > 0 {
			fmt.Println("\nStaged files:")
			fmt.Println(string(stagedFiles))
		}

		// Obtener el título del commit (primera línea del diff)
		title := strings.Split(string(diff), "\n")[0]

		// Generar descripción con ChatGPT
		description, err := generateCommitMessage(string(diff), title, apiKey)
		if err != nil {
			fmt.Println("Error generating commit message:", err)
			return
		}

		// Imprimir la descripción generada
		fmt.Println("\nGenerated commit message:")
		fmt.Println(description)

		// Preguntar al usuario si desea continuar
		fmt.Print("\nDo you want to proceed with git add . and commit? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled.")
			return
		}

		// Ejecutar git add .
		cmd2 := exec.Command("git", "add", ".")
		err = cmd2.Run()
		if err != nil {
			fmt.Println("Error executing git add .:", err)
			return
		}

		// Ejecutar git commit -m
		cmd2 = exec.Command("git", "commit", "-m", description)
		err = cmd2.Run()
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

func generateCommitMessage(diff, title, apiKey string) (string, error) {
	prompt := fmt.Sprintf("Based on the following git diff and the context '%s', generate a concise commit message in English. Start with a brief summary line (50 characters or less), followed by a blank line, and then a more detailed explanation. Do not include any prefixes like 'title:' or 'changes:'. Here's the diff:\n\n%s", title, diff)
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
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
