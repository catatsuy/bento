# bento

🍱

benri translator and optimizer

## Example

```.gitconfig
[alias]
sb = !git diff | bento -branch -model "gpt-4-turbo"
sc = !git diff --staged | bento -commit -model "gpt-4-turbo"
```
