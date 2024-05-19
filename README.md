# bento

🍱

benri translator and optimizer

## Example

```.gitconfig
[alias]
sb = !git diff -w | bento -branch -model "gpt-4o"
sc = !git diff -w --staged | bento -commit -model "gpt-4o"
```

To show new files in git diff, use the `git add -N` command. This stages the new files without adding content.

```bash
git add -N .
```

## Name Origin of "bento" 🍱

The name "bento" stands for **Benri Translator and Optimizer**. The word "benri" is Japanese for "convenient," which reflects the tool's purpose to provide handy solutions and optimizations.

In Japanese, "bento" refers to a lunch box that contains a variety of different dishes. This analogy is perfect for our tool because it combines multiple functionalities into one compact tool, much like a bento box 🍱 that offers a variety of foods in one compact container. Thus, our "bento" tool is designed to be a versatile and efficient assistant in your workflow, packing numerous features in an organized manner.
