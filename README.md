# work-env
Work-env is a command line environment for developers. It runs in Docker. Not clear, let me explain...

### The problem
When working on a software project, we need a lot of compilers, runtimes, interpreters, tools and libraries. Clasical way to get this packages is install it on your host system. It causes some problems:
* A version of a package is not available in your distribution. Distribution update or building from sources is painful.
* Conflicting versions. You work on multiple projects. The projects require different versions of the same tool.
* Build failures. Your collegue managed to build a project but it won't build on your machine.
* Non-reproducible builds. You managed to build a project but don't know which exacly dependencies where involved.
* Broken host system. You have installed something non-stable or non-compatible to your OS.

A lot of pain. **Never ever install project dependencies to your system.** Idealy you need an environment which is:
* Isolated. One project - one environment. Host system is not modified.
* Reproducible. You know for sure what dependencies are necessary and can share or recreate your environment.
* Universal. Supports any programming languages and technologies.
* Convenient. Working in environment must be as easy as in host system.

Some existing solutions:
* Virtual machines. You have a virtual machine per a project. Perfect isolation, heavyweight and inconvenient.
* Python Virtualenv. For Python only. Isolated set of libraries and Python tools. But depends on host interpreter and system libraries.
* Ruby RVM. For Ruby only. Like Virtualenv, but contains an interpreter in the envinronment.
* Linux system in chroot. Old school Linux solution. Whole system packed in a huge tarbal. Isolates all, but distribution is very inconvenient.
* Docker container. Perfect isolation. Very good solution for CI and deployment, but working interactively is rather inconvenient.

### The solution
Work-env is an environment which runs in Docker. Docker CLI and popular images where created to deploy services. Work-env is a CLI and a base image focused on interactive work.

-- Hey, I've got the idea. Do I need it if I'm good with Docker?

-- Sure you do. Because work-env:
* Has better CLI experience. Even for Docker gurus.
* Allows you to work as non-root user.
* Can run GUI applications.
* Automatically mounts your home directory.
* And has some other small but useful tweaks over default Docker behavior.


|                   | Host OS   | Virtual machine   | Python virtualenv | Ruby RVM  | Docker container  | Work-env  |
| ----------------- | --------  | ----------------- | ----------------- | --------- | ----------------- | --------  |
| Isolated libs     | no        | yes               | yes               | yes       | yes               | yes       |
| Isolated runtime  | no        | yes               | no                | yes       | yes               | yes       |
| Reproducible      | no        | yes               | yes               | yes       | yes               | yes       |
| Universal         | yes       | yes               | no                | no        | yes               | yes       |
| Convenient        | yes       | no                | yes               | yes       | no                | **yes**   |


### Protip
To make it easier to distinguish when you are working in a work-env, add next lines to your `~/.bashrc` or `~/.zshrc`:

```bash
if [ -n "$WORK_ENV_NAME" ]; then
    PS1=[${WORK_ENV_NAME}]${PS1}
fi
```
Now your shell prompt will have environment name at the beginning when you are working in an environment. Some terminal emulators also show environment hostname on a tab title.
