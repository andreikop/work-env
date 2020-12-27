# work-env


### Protip
To make it easier to distinguish when you are working in a work-env, add next lines to your `~/.bashrc` or `~/.zshrc`:

```bash
if [ -n "$WORK_ENV_NAME" ]; then
    PS1=[${WORK_ENV_NAME}]${PS1}
fi
```
Now your shell prompt will have environment name at the beginning when you are working in an environment. Some terminal emulators also show environment hostname on a tab title.
