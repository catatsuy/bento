# bento

🍱

bento is a CLI tool that uses OpenAI's API to assist with everyday tasks. It is especially useful for suggesting Git branch names, commit messages, and translating text.

## Features

- Uses **OpenAI's API** to assist with tasks.
- Easy-to-use commands: `-branch`, `-commit`, `-translate`.
- Supports **multi mode** and **single mode**:
  - **Single Mode**: Sends one request to the API. Used for `-branch` and `-commit`.
  - **Multi Mode**: Sends multiple requests to the API. Used for `-translate`.

## Name Origin of "bento" 🍱

The name "bento" stands for **Benri Translator and Optimizer**. The word "benri" is Japanese for "convenient," which reflects the tool's purpose to provide handy solutions and optimizations.

In Japanese, "bento" refers to a lunch box that contains a variety of different dishes. This analogy is perfect for our tool because it combines multiple functionalities into one compact tool, much like a bento box 🍱 that offers a variety of foods in one compact container. Thus, our "bento" tool is designed to be a versatile and efficient assistant in your workflow, packing numerous features in an organized manner.

## Installation

It is recommended that you use the binaries available on [GitHub Releases](https://github.com/catatsuy/bento/releases). It is advisable to download and use the latest version.

If you have a Go language development environment set up, you can also compile and install the 'bento' tools on your own.

```bash
go install github.com/catatsuy/bento@latest
```

To build and modify the 'bento' tools for development purposes, you can use the `make` command.

```bash
make
```

If you use the `make` command to build and install the 'bento' tool, the output of the `bento -version` command will be the git commit ID of the current version.

## Additional Information

- **API Token**: The API token is passed via the environment variable `OPENAI_API_KEY`.
- **Customization**: To customize, use `-multi` or `-single` and provide a custom prompt with `-prompt`.
- **Default Model**: The default model is `gpt-3.5-turbo`, but you can change it with the `-model` option.
- **Translation**: The `-translate` command translates to English by default; use `-language` to specify the target language.
- **File Handling**: To work with files, provide the filename with `-file` or use standard input.

## Usage Examples

```
Usage of bento:
  -branch
        Suggest branch name
  -commit
        Suggest commit message
  -file string
        specify a target file
  -h    Print help information and quit
  -help
        Print help information and quit
  -language string
        Translate to language (default: en)
  -limit int
        Limit the number of characters to translate (default 4000)
  -model string
        Use models such as gpt-3.5-turbo, gpt-4-turbo, and gpt-4o (default "gpt-3.5-turbo")
  -multi
        Multi mode
  -prompt string
        Prompt text
  -single
        Single mode
  -translate
        Translate text
  -version
        Print version information and quit
```

### Using `-branch` and `-commit`

- **`-branch`**: Use this when you haven't created a branch yet. It suggests a branch name based on the current Git diff.
  - Large new files can be problematic for the API to handle. By default, Git diff excludes new files, which is convenient. If necessary, add new files with `git add -N`.
- **`-commit`**: Use this when you are ready to commit. It suggests a commit message based on the staged files.
  - If new files cause large diffs, generate the commit message before staging them to avoid exceeding API limits.

Here is an example of setting up bento as a Git alias on `~/.gitconfig`. This allows you to generate branch names and commit messages from Git diffs automatically.

```.gitconfig
[alias]
sb = !git diff -w | bento -branch -model "gpt-4o"
sc = !git diff -w --staged | bento -commit -model "gpt-4o"
```

To show new files in git diff, use the `git add -N` command. This stages the new files without adding content.

```bash
git add -N .
```

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

The `-single` option is used when you need to send a single request to the API. This is useful for tasks that must be processed as a whole.

To summarize text from a file named example.txt, use the following command:

```sh
bento -single -prompt 'Please summarize the following text:\n\n' -file example.txt
```

## Tips

- **Prompt Tips**: When using `-branch`, the default prompt is: `Generate a branch name directly from the provided source code differences without any additional text or formatting:`. If you provide a custom prompt, it is recommended to add `"without any additional text or formatting"` at the end for better results.
