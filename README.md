# work-env
Work-env is a virtual command line environment for developers. It runs in Docker.

### The problem
When working on a software project, we need a lot of compilers, runtimes, interpreters, tools and libraries. Clasical way to get this packages is install it on your host system. It causes some problems:
* A version of a package you need is not available in your distribution. Distribution update or building the package from sources is painful.
* Conflicting versions. You work on multiple projects. The projects require different versions of the same library or tool.
* Build failures. Your collegue managed to build a project but it won't build on your machine because environment differs.
* Non-reproducible builds. You managed to build a project but don't know which exacly dependencies where involved.
* Broken host system. You have installed something non-stable or non-compatible to your OS.

A lot of pain. **Never ever install project dependencies to your OS.** Idealy you need an environment which is:
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

```
-- Hey, I understand the idea. But do I need work-env if I'm good with Docker?
-- Sure you do. Because work-env:
```

* Has better CLI experience. Even for Docker gurus.
* Supports working as non-root user.
* Can run GUI applications.
* Automatically mounts your home directory.
* And has some other small but useful tweaks over default Docker behavior.


work-env vs other isolated environment solutions:

|                   | Host OS   | Virtual machine   | Python virtualenv | Ruby RVM  | Docker container  | Work-env  |
| ----------------- | --------  | ----------------- | ----------------- | --------- | ----------------- | --------  |
| Isolated libs     | no        | yes               | yes               | yes       | yes               | yes       |
| Isolated runtime  | no        | yes               | no                | yes       | yes               | yes       |
| Reproducible      | no        | yes               | yes               | yes       | yes               | yes       |
| Universal         | yes       | yes               | no                | no        | yes               | yes       |
| Convenient        | yes       | no                | yes               | yes       | no                | **yes**   |

### Usage
#### One-time disposable environment
The simples use-case. You need an environment to do some experiment or build something once.
```bash
> work-env run --rm
Welcome to work-env. Your sudo password is 'ak'
> ./do-some-job.sh
> Ctrl^D
```

#### Create an environment for your project
More advanced case. You need an environment to work on one of your projects.
Using `work-env` as base image. This is a default Ubuntu-based image prepared by the team.

First run:
```bash
> work-env run work-env my-project
Welcome to work-env. Your sudo password is 'ak'
# Your work-env is ready. Now you can work here
> apt update
> apt install make
> make all
> ./deploy.sh
```

Tomorrow. An environment already exists. You can enter (attach) it and continue you job:
```bash
> work-env enter my-project
Welcome to work-env. Your sudo password is 'ak'
#continue working here
> git pull
> make all
> ./deploy.sh
```

#### Create standard environment for your team
Standard shared environment. A team of developers works on a single project. You need standard environment to make sure all problems and builds are reprodusible. New team members will be able to start working on the project without long and complicated process of installing required tools.

Create an environment for a team:
```bash
> vim Dockerfile
# Create a image, which contains your dependencies. You can use work-env as a base
> work-env build . example.com/my-team-env  # example.com is your team Docker repository
> docker push example.com/my-team-env  # Push built image so your team mates can pull it
```

Use the environment. On your collegues machine:
```bash
> work-env run example.com/my-team-env the-team-project
# Now your collegue works in environment prepared by you.
```

### How it works
work-env creates a Docker container per environment.

The container can be based on any interactive image, default is Ubuntu-based `work-env`.

When container is created, home directory is mounted to it, host working directory is preserved. Host networking is used.

Default base image `work-env` contains startup script to create non-root user whoes username and id equals to host user.

### Protip
To make it easier to distinguish when you are working in a work-env, add next lines to your host `~/.bashrc` or `~/.zshrc`:

```bash
if [ -n "$WORK_ENV_NAME" ]; then
    PS1=[${WORK_ENV_NAME}]${PS1}
fi
```
Now your shell prompt will have environment name at the beginning when you are working in an environment. Some terminal emulators also show environment hostname on a tab title.
