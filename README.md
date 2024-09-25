# git-ai

git-ai is a command-line tool that uses ChatGPT to generate commit messages for your Git repositories. It streamlines the process of creating meaningful and descriptive commit messages by analyzing your changes and generating a commit message based on the diff.

## Features

- Generates commit messages using OpenAI's GPT model
- Seamless integration with your Git workflow
- Easy configuration and setup

## Installation

### Option 1: Using the pre-compiled binary

If you don't have Go installed on your system, you can use the pre-compiled binary:

1. Go to the [Releases](https://github.com/yourusername/git-ai/releases) page of this repository.
2. Download the latest release for your operating system (Windows, macOS, or Linux).
3. Extract the downloaded archive.
4. Move the `git-ai` executable to a directory in your system's PATH. For example:

   - On macOS and Linux:
     ```
     sudo mv git-ai /usr/local/bin/
     ```
   - On Windows: 
     Move the `git-ai.exe` to a directory in your PATH, or add the directory containing `git-ai.exe` to your PATH environment variable.

5. Verify the installation by running:
   ```
   git-ai --version
   ```

### Option 2: Building from source

If you have Go 1.16 or higher installed, you can build the application from source:

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/git-ai.git
   ```

2. Navigate to the project directory:
   ```
   cd git-ai
   ```

3. Build the application:
   ```
   go build -o git-ai
   ```

4. Move the binary to a location in your PATH. For example:
   ```
   sudo mv git-ai /usr/local/bin/
   ```

### Prerequisites

- Git
- An OpenAI API key

## Configuration

Before using git-ai, you need to configure it with your OpenAI API key:

1. Run the configuration command:
   ```
   git-ai config
   ```

2. When prompted, enter your OpenAI API key.

The API key will be securely stored in `~/.git-ai/config.json`.

## Usage

### Generating a commit

To generate a commit message and create a commit:

1. Make changes to your Git repository as usual.

2. When you're ready to commit, run:
   ```
   git-ai commit
   ```

3. git-ai will analyze your changes and generate a commit message.

4. Review the generated message and confirm if you want to proceed with the commit.

5. If you confirm, git-ai will stage all changes (`git add .`) and create a commit with the generated message.

## Commands

- `git-ai config`: Configure the OpenAI API key
- `git-ai commit`: Generate a commit message and create a commit

## Notes

- git-ai uses `git add .` to stage all changes before committing. Make sure this aligns with your workflow.
- The generated commit messages are in English.
- Always review the generated commit message before confirming to ensure it accurately represents your changes.

## Troubleshooting

- If you encounter any issues with authentication, run `git-ai config` again to update your API key.
- Ensure you have an active internet connection, as git-ai needs to communicate with the OpenAI API.
- If you're using the pre-compiled binary and encounter a "command not found" error, make sure the directory containing the git-ai executable is in your system's PATH.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)