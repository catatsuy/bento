# bento

🍱
bento is a CLI tool that uses AI APIs to assist with everyday tasks. By default it uses OpenAI's API, but you can switch to Gemini's API with the `-backend` flag. It is especially useful for suggesting Git branch names, commit messages, translating text, and extracting repository contents.

## Features

- Uses **OpenAI's API** by default; support for **Gemini's API** is available via the `-backend gemini` flag.
- Extracts repository content with the `-dump` command.
- Easy-to-use commands: `-branch`, `-commit`, `-translate`, `-review`, and `-dump`.
- Supports **multi mode** and **single mode**:
  - **Single Mode**: Sends one request to the API (used for `-branch`, `-commit`, and `-review`).
  - **Multi Mode**: Sends multiple requests to the API (used for `-translate`).

## Name Origin of "bento" 🍱

The name "bento" stands for **Bundled ENhancements for Tasks and Operations**. This means the tool provides many useful features for different tasks, making it a helpful and flexible solution for various needs.

In Japanese, "bento" refers to a lunch box that contains a variety of different dishes. This analogy is perfect for our tool because it combines multiple functionalities into one compact tool, much like a bento box 🍱 that offers a variety of foods in one compact container. Thus, our "bento" tool is designed to be a versatile and efficient assistant in your workflow, packing numerous features in an organized manner.

## Installation

It is recommended that you use the binaries available on [GitHub Releases](https://github.com/catatsuy/bento/releases). Download and use the latest version.

Alternatively, if you have Go installed, compile and install bento with:

```bash
go install github.com/catatsuy/bento@latest
```

To build for development, use:

```bash
make
```

*(When built via `make`, `bento -version` outputs the current git commit ID.)*

## Additional Information

- **API Token**:
  - OpenAI: passed via the environment variable `OPENAI_API_KEY`.
  - Gemini: use the `-backend gemini` flag and set the token via `GEMINI_API_KEY`.
- **Repository Dump**: The `-dump` command extracts repository content while respecting `.gitignore` and `.aiignore`.
- **Customization**: Use `-multi` or `-single` and override prompts with `-prompt`.
- **Default Model**:
  - For OpenAI: default is `gpt-4o-mini`.
  - For Gemini (with `-backend gemini`): default is `gemini-2.0-flash-lite`.
- **Translation**: The `-translate` command translates to English by default; change target language with `-language`.
- **Code Review**: Use `-review` to get code feedback. Specify the output language with `-language`.
- **File Handling**: Provide a filename with `-file` or use standard input.

## Usage Examples

```
Usage of bento:
  -backend string
        Backend to use: openai or gemini (default "openai")
  -branch
        Suggest branch name
  -commit
        Suggest commit message
  -description string
        Description of the repository (dump mode)
  -dump
        Dump repository contents
  -file string
        Specify a target file
  -h    Print help information and quit
  -help
        Print help information and quit
  -language string
        Specify the output language
  -limit int
        Limit the number of characters to translate (default 4000)
  -model string
        Use models such as gpt-4o-mini, gpt-4-turbo, and gpt-4o. (When using the gemini backend, the default model becomes gemini-2.0-flash-lite) (default "gpt-4o-mini")
  -multi
        Multi mode
  -prompt string
        Prompt text
  -review
        Review source code
  -single
        Single mode (default)
  -system string
        System prompt text
  -translate
        Translate text
  -version
        Print version information and quit
```

### Using `-dump`

The `-dump` command is used to extract the contents of a Git repository in a structured format. Binary files are excluded, and `.gitignore` and `.aiignore` rules are respected.

To dump the contents of a repository, use:

```bash
bento -dump /path/to/repo
```

The output will follow this format:

1. Each section begins with `----`.
2. The first line after `----` contains the file path and name.
3. The subsequent lines contain the file contents.
4. The repository content ends with `--END--`.

#### Description Flag

The `-description` flag allows you to provide a specific description of the repository when using the dump mode. This description will be included in the output.

```bash
bento -dump -description "This is a sample repository description."
```

### Using `-branch` and `-commit`

- **`-branch`**: Use this when you haven't created a branch yet. It suggests a branch name based on the current Git diff.
  - Large new files can be problematic for the API to handle. By default, Git diff excludes new files, which is convenient. If necessary, add new files with `git add -N`.
- **`-commit`**: Use this when you are ready to commit. It suggests a commit message based on the staged files.
  - If new files cause large diffs, generate the commit message before staging them to avoid exceeding API limits.

Here is an example of setting up bento as a Git alias on `~/.gitconfig`. This allows you to generate branch names and commit messages from Git diffs automatically.

```.gitconfig
[alias]
sb = !git diff -w | bento -branch
sc = !git diff -w --staged | bento -commit
```

To show new files in git diff, use the `git add -N` command. This stages the new files without adding content.

```bash
git add -N .
```

### Using Review Mode with `-review`

The `-review` option is used when you need to review the source code. This mode focuses on identifying issues in various aspects such as Completeness, Bugs, Security, Code Style, etc.

You can also use the `-language` option to specify the output language of the review results.

To review code, use the following command:

```sh
git diff -w | bento -review -model gpt-4o -language Japanese
```

In this example, the review results will be in Japanese. You can change the output language by specifying a different language with `-language`.

For automation, you can use a GitHub Actions workflow. Below is an example workflow configuration file, [`.github/workflows/auto-review.yml`](/.github/workflows/auto-review.yml), which automatically runs a code review whenever there is a new pull request:

This workflow will trigger on every pull request and run a code review using the `bento` tool.

#### Important Security Considerations

- **API Key Storage**: Store your OpenAI API key as a repository secret named `OPENAI_API_KEY`:
  1. Go to your repository's Settings.
  2. Navigate to Secrets and variables > Actions.
  3. Click "New repository secret" to add the key.

- **Permissions**: Set proper permissions for Actions in open-source projects:
  1. In the repository settings, adjust the Actions permissions to "Allow OWNER, and select non-OWNER, actions and reusable workflows".
  2. For details, refer to the GitHub documentation [here](https://docs.github.com/github/administering-a-repository/disabling-or-limiting-github-actions-for-a-repository#allowing-select-actions-and-reusable-workflows-to-run).

### Using System Prompt with `-system`

The `-system` option allows you to define a system prompt text. This can be useful for customizing the initial instructions.

### Using `-translate`

The `-translate` option allows you to translate text to a target language. You can specify the target language using the `-language` option. By default, the target language is English (`en`).

#### Translating Text from a File

To translate text from a file named `example.txt` to Japanese, use the following command:

```sh
bento -translate -file example.txt -language ja
```

#### Translating Text from Standard Input

To translate text from standard input to French, use the following command.

```sh
echo 'hello' | bento -translate -language fr
```

### Using Multi Mode with `-multi`

To proofread a text and correct obvious errors while maintaining the original meaning and tone, use the following command:

```sh
bento -file textfile.txt -multi -prompt "Please correct only the obvious errors in the following text while maintaining the original meaning and tone as much as possible:\n\n"
```

### Using Single Mode with `-single`

The Single Mode is default. You don't need to specify `-single`.

The `-single` option is used when you need to send a single request to the API. This is useful for tasks that must be processed as a whole.

To summarize text from a file named example.txt, use the following command:

```sh
bento -single -prompt 'Please summarize the following text:\n\n' -file example.txt
```

## Tips

- **Prompt Suggestions**: The default prompts are optimized to produce minimal extra text. If using custom prompts, consider appending "without any additional text or formatting".
- **Backend Switching**: Use the `-backend` flag to switch between OpenAI and Gemini (token is provided via the corresponding environment variable).
- **Git Integration**: Set up Git aliases as shown above to generate branch names or commit messages directly from diffs.
