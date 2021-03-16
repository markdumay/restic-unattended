# restic-unattended

<!-- Tagline -->
<p align="center">
    <b>Schedule Unattended Remote Backups Using a Hardened Restic Container</b>
    <br />
</p>

<!-- Badges -->
<p align="center">
    <a href="https://github.com/markdumay/restic-unattended/actions/workflows/build-test.yml" alt="Build">
        <img src="https://img.shields.io/github/workflow/status/markdumay/restic-unattended/build.svg" />
    </a>
    <a href="https://github.com/markdumay/restic-unattended/actions/workflows/push.yml" alt="Docker Hub">
        <img src="https://img.shields.io/github/workflow/status/markdumay/restic-unattended/push.svg?label=docker" />
    </a>
    <a href="https://www.codefactor.io/repository/github/markdumay/restic-unattended" alt="CodeFactor">
        <img src="https://img.shields.io/codefactor/grade/github/markdumay/restic-unattended" />
    </a>
    <a href="https://github.com/markdumay/restic-unattended/commits/main" alt="Last commit">
        <img src="https://img.shields.io/github/last-commit/markdumay/restic-unattended.svg" />
    </a>
    <a href="https://github.com/markdumay/restic-unattended/issues" alt="Issues">
        <img src="https://img.shields.io/github/issues/markdumay/restic-unattended.svg" />
    </a>
    <a href="https://github.com/markdumay/restic-unattended/pulls" alt="Pulls">
        <img src="https://img.shields.io/github/issues-pr-raw/markdumay/restic-unattended.svg" />
    </a>
    <a href="https://github.com/markdumay/restic-unattended/blob/main/LICENSE" alt="License">
        <img src="https://img.shields.io/github/license/markdumay/restic-unattended.svg" />
    </a>
</p>

<!-- Table of Contents -->
<p align="center">
  <a href="#about">About</a> •
  <a href="#built-with">Built With</a> •
  <a href="#prerequisites">Prerequisites</a> •
  <a href="#testing">Testing</a> •
  <a href="#deployment">Deployment</a> •
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#credits">Credits</a> •
  <a href="#donate">Donate</a> •
  <a href="#license">License</a>
</p>


## About
[Restic][restic_url] is a fast and secure backup program. It supports  many backends for storing backups natively, including AWS S3, Openstack Swift, Backblaze B2, Microsoft Azure Blob Storage, and Google Cloud Storage. *Restic-unattended* is a helper utility written in Go to run automated restic backups using a built-in scheduler. Running as an unprivileged and hardened Docker container, *restic-unattended* simplifies the management of credentials and other sensitive data by using Docker secrets.

> **Looking for testers.** *Restic-unattended* has been integration tested with Backblaze B2. Your feedback on the integration with any other supported backend is much appreciated.

<!-- TODO: add tutorial deep-link 
Detailed background information is available on the author's [personal blog][blog].
-->

## Built With
The project uses the following core software components:
* [Cobra][cobra_url] - Go library to generate CLI applications (including [Viper][viper_url] and [pflag][pflag_url])
* [Cron][cron_url] - Go library to schedule jobs using cron notation
* [Dbm][dbm_url] - Helper utility to build, harden, and deploy Docker images
* [Docker][docker_url] - Open-source container platform
* [Restic][restic_url] - Secure backup program

## Prerequisites
*Restic-unattended* can run on any Docker-capable host. The setup has been tested locally on macOS Big Sur and in production on a server running Ubuntu 20.04 LTS. Cloud storage has been tested with Backblaze B2, although other storage providers are supported too.

* **Docker Engine and Docker Compose are required** - *restic-unattended* is intended to be deployed as a Docker container using Docker Compose for convenience. Docker Swarm is a prerequisite to enable Docker *secrets*, however, the use of Docker secrets itself is optional. This [reference guide][swarm_init] explains how to initialize Docker Swarm on your host.

* **A storage provider is required** - Restic supports several storage providers out of the box. Cloud providers include Amazon S3, Minio Server, Wasabi, OpenStack Swift, Backblaze B2, Microsoft Azure Blob Storage, and Google Cloud Storage. Next to that, local backups are supported too, as well as storage via SFTP, a REST server, or rclone. See the [restic documentation][restic_repo] for more details.

## Testing
It is recommended to test the services locally before deploying them to a production environment. Below four steps enable you to run the services on your local machine and validate they are working correctly. The configuration examples use Backblaze B2 as a storage provider. Check the [restic documentation][restic_repo] for the configuration of other storage providers. 

### Step 1 - Clone the Repository and Setup the Build Tool
The first step is to clone the repository to a local folder. Assuming you are in the working folder of your choice, clone the repository files with `git clone`. Git automatically creates a new folder `restic-unattended` and copies the files to this directory. Change your working folder to be prepared for the next steps.

```console
local:~$ git clone --recurse-submodules https://github.com/markdumay/restic-unattended.git
local:~$ cd restic-unattended
```

The repository uses [dbm][dbm_url] to simplify the build and deployment process. Set up an alias to simplify the execution of dbm.
```console
local:~/restic-unattended$ alias dbm="dbm/dbm.sh"  
```
Add the same line to your shell settings (e.g. `~/.zshrc` on macOS or `~/.bashrc` on Ubuntu with bash login) to make the alias persistent.

### Step 2 - Configure the Environment Variables
The `docker-compose.yml` file uses environment variables to simplify the configuration. You can use the sample file in the repository as a starting point. The provided sample configuration applies to the production environment.

```console
local:~/restic-unattended$ mv sample.env .env
```

### Step 3 - Specify the Storage Provider Credentials
Pending on your selected storage provider, you will need to specify the tokens and/or account credentials for restic to be able to connect with the provider. You can either specify the credentials as environment variables or as Docker secrets. In this example, we will use Docker secrets. See the project Wiki for a detailed overview of both [configuration options][wiki_local].

As regular Docker containers do not support external Swarm secrets, we will create local secret files for testing purposes. The credentials are stored in plain text, so this is not recommended for production. Add the secrets to `docker/docker-compose.yml` and `docker/docker-compose.dev.yml` to provide the credentials for the `restic-unattended` container. The `docker-compose.dev.yml` extends the base file `docker-compose.yml` to simplify the debugging, building, and running of Docker images.

> Mounting file-based secrets from your working directory does not work with user namespaces enabled. See the [documentation][docker_userns] for more details.

#### Define the Base Configuration
Ensure the following configuration settings are defined in `docker-compose.yml`. Please note that the sensitive environment variables now have a `_FILE` suffix and point to the location `/run/secrets/`. Docker mounts secrets to this location within the container by default.

```yml
version: "3.7"

secrets:
  RESTIC_REPOSITORY:
    external: true
  RESTIC_PASSWORD:
    external: true
  B2_ACCOUNT_ID:
    external: true
  B2_ACCOUNT_KEY:
    external: true
[...]

services:
  restic:
    [...]
    environment:
      - RESTIC_REPOSITORY_FILE=/run/secrets/RESTIC_REPOSITORY
      - RESTIC_PASSWORD_FILE=/run/secrets/RESTIC_PASSWORD
      - B2_ACCOUNT_ID_FILE=/run/secrets/B2_ACCOUNT_ID
      - B2_ACCOUNT_KEY_FILE=/run/secrets/B2_ACCOUNT_KEY
    secrets:
      - RESTIC_REPOSITORY
      - RESTIC_PASSWORD
      - B2_ACCOUNT_ID
      - B2_ACCOUNT_KEY
```

#### Define the Development Configuration
The development configuration overrides the base configuration defined in the previous section. Point to the file-based secrets with the below configuration in `docker-compose.dev.yml`.

```yml
secrets:
  RESTIC_REPOSITORY:
    file: secrets/RESTIC_REPOSITORY
    external: false
  RESTIC_PASSWORD:
    file: secrets/RESTIC_PASSWORD
    external: false
  B2_ACCOUNT_ID:
    file: secrets/B2_ACCOUNT_ID
    external: false
  B2_ACCOUNT_KEY:
    file: secrets/B2_ACCOUNT_KEY
    external: false
```

#### Create the File-based Docker Secrets
The final step is to create the file-based secrets themselves. Replace the `XXX` values with your credentials. 

```console
local:~/restic-unattended$ mkdir secrets
local:~/restic-unattended$ printf XXXXX > secrets/RESTIC_REPOSITORY
local:~/restic-unattended$ printf XXXXX > secrets/RESTIC_PASSWORD
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXX > secrets/B2_ACCOUNT_ID
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX > secrets/B2_ACCOUNT_KEY
```

### Step 4 - Run the Docker Container
The repository contains a helper script to run the Docker container. Use the below commands to build and run an image with built-in shell support for debugging.
```console
local:~/restic-unattended$ dbm dev build
local:~/restic-unattended$ dbm dev up -t
```

The script then validates the host machine, identifies the targeted image, and brings up the container and network. You should now be logged in to the container's shell. Try a few commands, such as `restic-unattended list` to verify everything is working as expected. See the Wiki for guidance on how to conduct a more elaborate [integration test][wiki_it] from within the container.

## Deployment
The steps for deploying in production are slightly different than for local testing. The next four steps highlight the changes.

### Step 1 - Clone the Repository
*Unchanged*

### Step 2 - Update the Environment Variables
The sample configuration writes log messages and errors with a timestamp to the standard output. It also defines several constraints following recommendations from the [Docker Bench for Security][docker_bench]. Finally, the production image defines a `RESTIC_CMD` to enable unattended backups. Ensure the `docker-compose.prod.yml` file captures the following settings.


```yml
services:
  restic:
    [...]
    environment:
      - RESTIC_LOGLEVEL=${RESTIC_LOGLEVEL}
      - RESTIC_LOGFORMAT=${RESTIC_LOGFORMAT}
      - RESTIC_TIMESTAMP=${RESTIC_TIMESTAMP}
    command: "${RESTIC_CMD}"

deploy:
      [...]
      resources:
        limits:
          cpus: "${RESTIC_LIMIT_CPU}"
          memory: "${RESTIC_LIMIT_MEM}"
        reservations:
          cpus: "${RESTIC_RESERVATION_CPU}"
          memory: "${RESTIC_RESERVATION_MEM}"
```

Adjust the logging settings and the deployment configuration for the CPU and allocated memory in the `.env` file as needed. The defined command runs a scheduled backup every 15 minutes and removes obsolete snapshots following a policy every day at 01:00 am.

```ini
RESTIC_LOGLEVEL=info
RESTIC_LOGFORMAT=pretty
RESTIC_LIMIT_CPU='0.25'
RESTIC_LIMIT_MEM='100M'
RESTIC_RESERVATION_CPU='0.05'
RESTIC_RESERVATION_MEM='6M'
RESTIC_CMD=restic-unattended schedule '0/15 * * * *' -p=/data/backup --forget='0 1 * * *' --keep-last=5 --keep-daily=7 --keep-weekly=13 --sustained
```

### Step 3 - Specify the Storage Provider Credentials
Instead of file-based secrets, you will now create more secure secrets. Docker secrets can be easily created using pipes. Do not forget to include the final `-`, as this instructs Docker to use piped input. Update the values as needed.

```console
local:~/restic-unattended$ printf XXXXX | docker secret create RESTIC_REPOSITORY -
local:~/restic-unattended$ printf XXXXX | docker secret create RESTIC_PASSWORD -
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXX | docker secret create B2_ACCOUNT_ID -
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX | docker secret create B2_ACCOUNT_KEY -
```

If you do not feel comfortable copying secrets from your command line, you can use the wrapper `create_secret.sh`. This script prompts for a secret and ensures sensitive data is not displayed on your console. The script is available in the folder `./docker-secret` of your repository.

```console
local:~/restic-unattended$ ./docker-secret/create_secret.sh RESTIC_REPOSITORY
local:~/restic-unattended$ ./docker-secret/create_secret.sh RESTIC_PASSWORD
local:~/restic-unattended$ ./docker-secret/create_secret.sh B2_ACCOUNT_ID
local:~/restic-unattended$ ./docker-secret/create_secret.sh B2_ACCOUNT_KEY
```

### Step 4 - Run the Docker Service
Docker Swarm is needed to support external Docker secrets. As such, the services will be deployed as a Docker stack in production. The helper script `dbm` generates the configuration using the applicable `.yml` files and deploys the services to the `restic` stack. 

```console
local:~/restic-unattended$ dbm prod build
local:~/restic-unattended$ dbm prod deploy
```

Check the status of the Docker service with the following command. The built-in health check kicks in after 5 minutes by default, by which the status should have changed from `starting` to `healthy`.

```console
local:~/restic-unattended$ docker ps
```

Run the following command to inspect the status of the Docker stack itself.

```console
local:~/restic-unattended$ docker stack services restic
```

You should see the value `1/1` for `REPLICAS` for the restic service if the stack was initialized correctly. You can view the service logs with `docker service logs restic_restic` once the service is up and running. Debugging Swarm services can be quite tedious. If for some reason your service does not initiate properly, you can get its task ID with `docker service ps restic_restic`. Running `docker inspect <task-id>` might give you clues to what is happening. Use `docker stack rm restic` to remove the Docker stack entirely.

## Usage
*Restic-unattended* is intended to run from within a Docker container as an unattended service. As such, the most common use case is to define a schedule command in the `docker/docker-compose.yml` file. In production, it is recommended to add the `--sustained` flag to ensure *restic-unattended* keeps running despite errors. Be sure to use the `[""]` notation for the `cmd` specification, as the production-ready image has no built-in shell.

*Restic-unattended* can also run from the command line (either from the host or from within a development container). The below list describes several scenarios.
* **From source (command line)** - the host machine requires Go 1.16 or later to be installed. From the `src` directory, run `go run main.go` followed by a specific command.
* **From source (Visual Studio Code)** - an `example-launch.json` file is included in the repository. Copy the configuration to `.vscode/launch.json` and set `env` and `args` as needed. The Go language tools need to be installed too.
* **From within a container** - spin up a development container with `dbm dev up -t`. Run `restic-unattended` from within the container.

Several commands and flags are supported, which are described in the project [Wiki][wiki_cmd]. They can also be inspected by using either `restic-unattended -h` or `restic-unattended <command> -h`.

## Contributing
1. Clone the repository and create a new branch 
    ```console
    local:~$ git checkout https://github.com/markdumay/restic-unattended.git -b name_for_new_branch
    ```
2. Make and test the changes
3. Submit a Pull Request with a comprehensive description of the changes

## Credits
The scheduler routine of *restic-unattended* is inspired by the following blog article:
* OpsDash by RapidLoop - [Job Queues in Go][go_queue]

## Donate
<a href="https://www.buymeacoffee.com/markdumay" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/lato-orange.png" alt="Buy Me A Coffee" style="height: 51px !important;width: 217px !important;"></a>

## License
The restic-unattended codebase is released under the [MIT license][license]. The README.md file, and files in the "[wiki][wiki]" repository are licensed under the Creative Commons *Attribution-NonCommercial 4.0 International* ([CC BY-NC 4.0)][cc-by-nc-4.0] license.


<!-- MARKDOWN PUBLIC LINKS -->
[cc-by-nc-4.0]: https://creativecommons.org/licenses/by-nc/4.0/
[cobra_url]: https://github.com/spf13/cobra
[cron_url]: https://github.com/robfig/cron/
[cron_usage]: https://godoc.org/github.com/robfig/cron#hdr-Usage
[docker_bench]: https://github.com/docker/docker-bench-security
[docker_userns]: https://docs.docker.com/engine/security/userns-remap/
[docker_url]: https://docker.com
[go_queue]: https://www.opsdash.com/blog/job-queues-in-go.html
[pflag_url]: https://github.com/spf13/pflag
[restic_url]: https://restic.net
[restic_repo]: https://restic.readthedocs.io/en/stable/030_preparing_a_new_repo.html
[restic_env]: https://restic.readthedocs.io/en/latest/040_backup.html
[swarm_init]: https://docs.docker.com/engine/reference/commandline/swarm_init/
[viper_url]: https://github.com/spf13/viper

<!-- MARKDOWN MAINTAINED LINKS -->
<!-- TODO: add blog link
[blog]: https://markdumay.com
-->
[blog]: https://github.com/markdumay
[repository]: https://github.com/markdumay/restic-unattended.git
[dbm_url]: https://github.com/markdumay/dbm.git
[license]: https://github.com/markdumay/restic-unattended/blob/main/LICENSE
[wiki]: https://github.com/markdumay/restic-unattended/wiki/
[wiki_cmd]: https://github.com/markdumay/restic-unattended/wiki/Available-Commands
[wiki_env]: https://github.com/markdumay/restic-unattended/wiki/Supported-Environment-Variables
[wiki_it]: https://github.com/markdumay/restic-unattended/wiki/Integration-Testing
[wiki_local]: https://github.com/markdumay/restic-unattended/wiki/Local-Testing