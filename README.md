# restic-unattended (work in progress)

<!-- Tagline -->
<p align="center">
    <b>Schedule Unattended Remote Backups Using a Hardened Restic Container</b>
    <br />
</p>


<!-- Badges -->
<p align="center">
    <a href="https://github.com/markdumay/restic-unattended/actions/workflows/buildtestlint.yml" alt="build">
        <img src="https://github.com/markdumay/restic-unattended/actions/workflows/buildtestlint.yml/badge.svg" />
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
[Restic][restic_url] is a fast and secure backup program. It supports  many backends for storing backups natively, including AWS S3, Openstack Swift, Backblaze B2, Microsoft Azure Blob Storage, and Google Cloud Storage. *Restic-unattended* is a helper utility written in Go to run restic backups with a built-in scheduler. Running as an unprivileged and hardened Docker container, *restic-unattended* simplifies the management of credentials and other sensitive data by using Docker secrets.

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
It is recommended to test the services locally before deploying them in a production environment. Running the services with `docker-compose` greatly simplifies validating everything is working as expected. Below four steps will allow you to run the services on your local machine and validate they are working correctly.

### Step 1 - Clone the Repository and Setup the Build Tool
The first step is to clone the repository to a local folder. Assuming you are in the working folder of your choice, clone the repository files with `git clone`. Git automatically creates a new folder `restic-unattended` and copies the files to this directory. Change your working folder to be prepared for the next steps. The code examples use Backblaze B2 as the storage provider. Be sure to replace the credentials with the correct ones.

```console
local:~$ git clone --recurse-submodules https://github.com/markdumay/restic-unattended.git
local:~$ cd restic-unattended
```

The repository uses [dbm][dbm_url] to simplify the build and deployment process. Setup an alias to simplify the execution of dbm.
```console
local:~/restic-unattended$ alias dbm="dbm/dbm.sh"  
```
Add the same line to your shell settings (e.g. ~/.zshrc on macOS or ~/.bashrc on Ubuntu with bash login) to make the alias persistent.

### Step 2 - Update the Environment Variables
The `docker-compose.yml` file uses environment variables to simplify the configuration. You can use the sample file in the repository as a starting point.

```console
local:~/restic-unattended$ mv sample.env .env
```

*Restic-unattended* recognizes the various [environment variables][restic_env] supported by restic. On top of that, several variables with the suffix `_FILE` are introduced to support Docker secrets. Lastly, three additional variables are supported to simplify configuration and logging. The below table gives an overview of the available environment variables.

| Variable                              | Secret | Description |
|---------------------------------------|--------|-------------|
| *Default restic variables*            |        | |
| RESTIC_REPOSITORY                     | *      | Location of repository (replaces -r) |
| RESTIC_REPOSITORY_FILE                | Yes    | Name of file containing the repository location |
| RESTIC_PASSWORD                       | *      | The actual password for the repository |
| RESTIC_PASSWORD_FILE                  | Yes    | Name of file containing the restic password |
| RESTIC_PASSWORD_COMMAND               |        | Command printing the password for the repository to stdout |
| RESTIC_KEY_HINT                       |        | ID of the key to try decrypting first, before other keys |
| RESTIC_CACHE_DIR                      |        | Location of the cache directory |
| RESTIC_PROGRESS_FPS                   |        | Frames per second by which the progress bar is updated |
| *Additional variables*                |        | |
| RESTIC_BACKUP_PATH                    |        | Local path to backup |
| RESTIC_HOST                           |        | Hostname to use in backups (defaults to `$HOSTNAME`) |
| RESTIC_LOGLEVEL                       |        | Level of logging to use: panic, fatal, error, warn, info, debug, trace |
| RESTIC_LOGFORMAT                      |        | Log format to use: default, pretty, json (schedule defaults to pretty) |
| RESTIC_TIMESTAMP                      |        | Add timestamp prefix (RFC3339) to logs: true, false |
| *Amazon S3 compatible*                |        | |
| AWS_ACCESS_KEY_ID                     | *      | Amazon S3 access key ID |
| AWS_ACCESS_KEY_ID_FILE                | Yes    | Name of file containing the Amazon S3 access key ID |
| AWS_SECRET_ACCESS_KEY                 | *      | Amazon S3 secret access key |
| AWS_SECRET_ACCESS_KEY_FILE            | Yes    | Name of file containing the Amazon S3 secret access key |
| AWS_DEFAULT_REGION                    |        | Amazon S3 default region |
| *OpenStack Swift (keystone v1)*       |        | |
| ST_AUTH                               |        | Auth URL for keystone v1 authentication |
| ST_USER                               | *      | Username for keystone v1 authentication |
| ST_USER_FILE                          | Yes    | Name of file containing the Username for keystone v1 authentication |
| ST_KEY                                | *      | Password for keystone v1 authentication |
| ST_KEY_FILE                           | Yes    | Name of file containing the Password for keystone v1 authentication |
| *OpenStack Swift (keystone v2/v3)*    |        | |
| OS_AUTH_URL                           |        | Auth URL for keystone authentication |
| OS_REGION_NAME                        |        | Region name for keystone authentication |
| OS_USERNAME                           | *      | Username for keystone authentication |
| OS_USERNAME_FILE                      | Yes    | Name of file containing the Username for keystone authentication |
| OS_PASSWORD                           | *      | Password for keystone authentication |
| OS_PASSWORD_FILE                      | Yes    | Name of file containing the Password for keystone authentication |
| OS_TENANT_ID                          | *      | Tenant ID for keystone v2 authentication |
| OS_TENANT_ID_FILE                     | Yes    | Name of file containing the Tenant ID for keystone v2 authentication |
| OS_TENANT_NAME                        | *      | Tenant name for keystone v2 authentication |
| OS_TENANT_NAME_FILE                   | Yes    | Name of file containing the Tenant name for keystone v2 authentication |
| *OpenStack Swift (keystone v3)*       |        | |
| OS_USER_DOMAIN_NAME                   | *      | User domain name for keystone authentication |
| OS_USER_DOMAIN_NAME_FILE              | Yes    | Name of file containing the User domain name for keystone authentication |
| OS_PROJECT_NAME                       | *      | Project name for keystone authentication |
| OS_PROJECT_NAME_FILE                  | Yes    | Name of file containing the Project name for keystone authentication |
| OS_PROJECT_DOMAIN_NAME                | *      | Project domain name for keystone authentication |
| OS_PROJECT_DOMAIN_NAME_FILE           | Yes    | Name of file containing the Project domain name for keystone authentication |
| *OpenStack Swift (credentials)*       |        | |
| OS_APPLICATION_CREDENTIAL_ID          | *      | Application Credential ID (keystone v3) |
| OS_APPLICATION_CREDENTIAL_ID_FILE     | Yes    | Name of file containing the Application Credential ID (keystone v3) |
| OS_APPLICATION_CREDENTIAL_NAME        | *      | Application Credential Name (keystone v3) |
| OS_APPLICATION_CREDENTIAL_NAME_FILE   | Yes    | Name of file containing the Application Credential Name (keystone v3) |
| OS_APPLICATION_CREDENTIAL_SECRET      | *      | Application Credential Secret (keystone v3) |
| OS_APPLICATION_CREDENTIAL_SECRET_FILE | Yes    | Name of file containing the Application Credential Secret (keystone v3) |
| *OpenStack Swift (token)*             |        | |
| OS_STORAGE_URL                        |        | Storage URL for token authentication |
| OS_AUTH_TOKEN                         | *      | Auth token for token authentication |
| OS_AUTH_TOKEN_FILE                    | Yes    | Name of file containing the Auth token for token authentication |
| *Backblaze B2*                        |        | |
| B2_ACCOUNT_ID                         | *      | Account ID or applicationKeyId for Backblaze B2 |
| B2_ACCOUNT_ID_FILE                    | Yes    | Name of file containing the Account ID or applicationKeyId for Backblaze B2 |
| B2_ACCOUNT_KEY                        | *      | Account Key or applicationKey for Backblaze B2 |
| B2_ACCOUNT_KEY_FILE                   | Yes    | Name of file containing the Account Key or applicationKey for Backblaze B2 |
| *Microsoft Azure Blob Storage*        |        | |
| AZURE_ACCOUNT_NAME                    | *      | Account name for Azure |
| AZURE_ACCOUNT_NAME_FILE               | Yes    | Name of file containing the Account name for Azure |
| AZURE_ACCOUNT_KEY                     | *      | Account key for Azure |
| AZURE_ACCOUNT_KEY_FILE                | Yes    | Name of file containing the Account key for Azure |
| *Google Cloud Storage*                |        | |
| GOOGLE_PROJECT_ID                     | *      | Project ID for Google Cloud Storage |
| GOOGLE_PROJECT_ID_FILE                | Yes    | Name of file containing the Project ID for Google Cloud Storage |
| GOOGLE_APPLICATION_CREDENTIALS        |        | Application Credentials for Google Cloud Storage (e.g. $HOME/.config/gs-secret-restic-key.json) |
| *rclone settings*                     |        | |
| RCLONE_BWLIMIT                        |        | rclone bandwidth limit |


### Step 3 - Specify Storage Provider Credentials
Pending on your selected storage provider, you will need to specify the tokens and/or account credentials for restic to able to connect with the provider. You can either specify the credentials as environment variables or as Docker secrets. Docker secrets are a bit more secure and are more suitable for a production environment. Please check the documentation of your storage provider in the <a href="#prerequisites">Prerequisites</a> section. 

#### Option 3a - Using Environment Variables
Backblaze B2 requires an account ID and account key to connect with the repository. Verify that the repository link in the `.env` file starts with the prefix `b2:`, followed by the correct repository name and `:/`. Finally, restic uses `RESTIC_PASSWORD` to encrypt your data. This means that the data cannot be viewed using the Backblaze web interface, but has to be restored using restic instead. Add the following lines to `docker/docker-compose.yml`:
```yml
[...]
services:
  restic:
    [...]
    environment:
      - RESTIC_REPOSITORY=${RESTIC_REPOSITORY}
      - RESTIC_PASSWORD=${RESTIC_PASSWORD}
      - B2_ACCOUNT_ID=${B2_ACCOUNT_ID}
      - B2_ACCOUNT_KEY=${B2_ACCOUNT_KEY}
```

Ensure the following variables are available in your `.env` file, replacing `XXX` with the real values:
```ini
# Restic settings
RESTIC_REPOSITORY=b2:XXXXX:/
RESTIC_PASSWORD=XXXXX

# Backblaze B2 credentials
B2_ACCOUNT_ID=XXXXXXXXXXXXXXXXXXXXXXXXX
B2_ACCOUNT_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

#### Option 3b - Using Docker Secrets
As Docker Compose does not support external Swarm secrets, we will create local secret files for testing purposes. The credentials are stored as plain text, so this is not recommended for production. Add the secrets to `docker/docker-compose.yml` and `docker/docker-compose.dev.yml` to authorize the restic service. The various docker-compose.yml files are chained in the order of `docker-compose.yml`, `docker-compose.prod.yml`, and `docker-compose.dev.yml` to simplify debugging, building, and running of images and containers. Each file overrides the settings of previous files in the chain.
```yml
version: "3"

secrets:
  RESTIC_PASSWORD:
    file: ../secrets/RESTIC_PASSWORD
  B2_ACCOUNT_ID:
    file: ../secrets/B2_ACCOUNT_ID
  B2_ACCOUNT_KEY:
    file: ../secrets/B2_ACCOUNT_KEY

[...]

services:
  restic:
    [...]
    environment:
      - RESTIC_REPOSITORY=${RESTIC_REPOSITORY}
      - RESTIC_PASSWORD_FILE=/run/secrets/RESTIC_PASSWORD
      - B2_ACCOUNT_ID_FILE=/run/secrets/B2_ACCOUNT_ID
      - B2_ACCOUNT_KEY_FILE=/run/secrets/B2_ACCOUNT_KEY
    secrets:
      - RESTIC_PASSWORD
      - B2_ACCOUNT_ID
      - B2_ACCOUNT_KEY
```

Ensure the following variable is available in your `.env` file, replacing `XXX` with the real value:
```ini
# Restic settings
RESTIC_REPOSITORY=b2:XXXXX:/
```

Please note that the sensitive environment variables now have a `_FILE` suffix and all point to the location `/run/secrets/`. Docker mounts all secrets to this location by default. Now create the following file-based secrets:
```console
local:~/restic-unattended$ mkdir secrets
local:~/restic-unattended$ printf XXXXX > secrets/RESTIC_PASSWORD
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXX > secrets/B2_ACCOUNT_ID
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX > secrets/B2_ACCOUNT_KEY
```

### Step 4 - Run Docker Container
The repository contains a helper script to run the Docker container. Use the below command to download an image for debugging (which has built-in shell support) and access the command line from within the container.
```console
local:~/restic-unattended$ dbm dev up -t
```

The script then validates the host machine, identifies the targeted image, and brings up the container and network. Running on a mac in the development mode `dev`, the output should look similar to this.
```console
Warn: env file not found
Validating environment
  Docker Engine: v20.10.0
  Docker Compose: v1.27.4
  Host: Darwin/amd64

Identifying targeted images
  restic_image="markdumay/restic-unprivileged:0.5.0-debug"

Bringing containers and networks up

Creating network "docker_restic" with the default driver
Creating docker_restic_1 ... done
```


#### Test the Environment Variables
From within the container, run the following command to validate the environment variables are properly initialized.
```console
container:~$ restic-unattended list
```

The output should look similar to this:
```console
VARIABLE            	SET	DESCRIPTION                                                                 
B2_ACCOUNT_ID_FILE  	Yes	Name of file containing the Account ID or applicationKeyId for Backblaze B2	
B2_ACCOUNT_KEY_FILE 	Yes	Name of file containing the Account Key or applicationKey for Backblaze B2 	
RESTIC_PASSWORD_FILE	Yes	Name of file containing the restic password                                	
RESTIC_REPOSITORY   	Yes	Location of the repository                                                     	
```

To review *all* supported environment variables instead, run `restic-unattended list -a`.

#### Making the First Backup
Now let's create some test data to test the backup functionality. Create a file `test.txt` in the folder `~/backup` and enter some sample text.

> The Docker volume `/data/backup` is mounted as read-only by default. Typically it is shared with another container that has full ownership of the volume. The UID and GID of the user running the container need to be the same.

```console
container:~$ mkdir -p ~/backup
container:~$ printf "This is a sample file to test restic backup and restore" > ~/backup/backup/test.txt
```

Now test the configuration by creating the first backup. Use the following command to create a backup on the spot and to init the remote repository if needed.
```console
container:~$ restic-unattended backup -p ~/backup --init
```

After some processing, the output should look similar to this.
```console
Starting backup operation of path '/home/restic/backup'
no parent snapshot found, will read all files

Files:           1 new,     0 changed,     0 unmodified
Dirs:            3 new,     0 changed,     0 unmodified
Added to the repo: 1.414 KiB

processed 1 files, 55 B in 0:07
snapshot xxxxxxxx saved
Finished backup operation of path '/home/restic/backup'
```

#### Restoring the First Backup
Now perform a restore operation to test if the backup did work. Use the following command to restore the data to the `/data/restore` folder, which has been created by Docker during initialization. Please note that restore uses the latest available snapshot by default, but can be instructed to use a specific snapshot instead. See `restic-unattended restore -h` and `restic-unattended snapshots -h` for more details.
```console
container:~$ restic-unattended restore /data/restore
```

The output should look similar to this.
```console
Starting restore operation for snapshot 'latest'
restoring <Snapshot XXXXXXXX of [/home/restic/backup] at DATE by USER> to /data/restore

Finished restore operation for snapshot 'latest'
```

Navigate to the `/data/restore` folder to verify the data has been restored correctly.


#### Scheduling an Unattended Backup
Test the scheduled backup functionality once the one-off backup is working correctly. *Restic-unattended* has a built-in scheduler capable of interpreting cron specifications. It supports several keywords and optional seconds too (emulating the alternative cron format provided by Quartz). See the [cron documentation][cron_usage] for more details. For testing purposes, we will instruct *restic-unattended* to perform a backup every minute at zero seconds, and a forget operation every minute at 30 seconds. The folder to backup is specified by the `-p` flag. It can also be configured as an environment variable instead. The forget operation instructs restic to remove old snapshots according to a policy. In this case, we instruct restic to keep the last 5 snapshots. See `restic-unattended forget -h` for more options.

The scheduler fires `backup` and `forget` jobs at the specified intervals. In this test, the data to be backed up is very small and the jobs are both expected to finish in less than 30 seconds. The scheduler processes one job at a time only, following a First In, First Out (FIFO) policy. If a current job is still running, the next job is delayed until the current job has finished. A maximum of 5 jobs is kept at any time. Additional jobs will be dropped when the maximum capacity has been reached. In practice, it is recommended to time the typical duration of your jobs and to set a realistic schedule for both types of jobs.

Test the scheduling functionality with the following command.

> Due to limitations of the used software libraries, the `~` expansion does not work when using `-p=~/backup`. Use a format without the `=` sign instead.

```console
container:~$ restic-unattended schedule '0 * * * * *' -p ~/backup --forget '30 * * * * *' --keep-last 5
```


The schedule job keeps on running until you hit `ctrl-c`. You should see logging output similar to the below example. The timestamps have been removed for brevity (and can also be omitted by using the flag `--logformat=default`). 

The first section displays several initialization messages.
```console
INFO   | Executing schedule command
INFO   | Scheduling job 'backup' with cron spec '0 * * * * *'
INFO   | First 'backup' job scheduled to run at 'TIME'
INFO   | Scheduling job 'forget' with cron spec '30 * * * * *'
INFO   | First 'forget' job scheduled to run at 'TIME'
```

The `forget` operation shows output similar to this once it started running.
```console
INFO   | Starting forget operation
INFO   | Applying Policy: keep 5 latest snapshots
INFO   | keep 3 snapshots:
INFO   | ID        Time  Host  Tags  Reasons        Paths
INFO   | -------------------------------------------------------
INFO   | XXXXXXXX  TIME  host        last snapshot  /data/backup
INFO   | YYYYYYYY  TIME  host        last snapshot  /data/backup
INFO   | ZZZZZZZZ  TIME  host        last snapshot  /data/backup
INFO   | -------------------------------------------------------
INFO   | 3 snapshots
INFO   | keep 4 snapshots:
INFO   | ID        Time  Host  Tags  Reasons        Paths
INFO   | -------------------------------------------------------
INFO   | WWWWWWWW  TIME  host        last snapshot  /data/backup
INFO   | XXXXXXXX  TIME  host        last snapshot  /data/backup
INFO   | YYYYYYYY  TIME  host        last snapshot  /data/backup
INFO   | ZZZZZZZZ  TIME  host        last snapshot  /data/backup
INFO   | -------------------------------------------------------
INFO   | 4 snapshots
INFO   | keep 1 snapshots:
INFO   | ID        Time  Host  Tags  Reasons        Paths
INFO   | -------------------------------------------------------
INFO   | ZZZZZZZZ  TIME  host        last snapshot  /data/backup
INFO   | -------------------------------------------------------
INFO   | 1 snapshots
INFO   | Finished forget operation
```

The `backup` operation shows output similar to this.
```console
INFO   | Starting backup operation of path '/home/restic/backup'
INFO   | Files:           0 new,     0 changed,     1 unmodified
INFO   | Dirs:            0 new,     2 changed,     0 unmodified
INFO   | Added to the repo: 342 B
INFO   | processed 1 files, 3.045 KiB in 0:03
INFO   | snapshot XXXXXXXX saved
INFO   | Finished backup operation of path '/home/restic/backup'
```

Hit `ctrl-c` to stop the scheduler.
```console
WARN   | Worker processing canceled
FATAL  | Error running schedule command error="Cron processing interrupted"
```

#### Stopping the Test
Stop execution of the container by entering `exit` on the container's command line. Dbm then automatically removes the created container and network.
```console
Stopping docker_restic_1 ... done
Removing docker_restic_1 ... done
Removing network docker_restic
Done.
```

## Deployment
The steps for deploying in production are slightly different than for local testing. The next four steps highlight the changes.

### Step 1 - Clone the Repository
*Unchanged*

### Step 2 - Update the Environment Variables
*Unchanged*

### Step 3 - Specify Storage Provider Credentials
#### Option 3a - Using Environment Variables
*Unchanged*

#### Option 3b - Using Docker Secrets
Instead of file-based secrets, you will now create more secure secrets. Docker secrets can be easily created using pipes. Do not forget to include the final `-`, as this instructs Docker to use piped input. Update the tokens as needed.

```console
local:~/restic-unattended$ printf XXXXX | docker secret create RESTIC_PASSWORD -
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXX | docker secret create B2_ACCOUNT_ID -
local:~/restic-unattended$ printf XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX | docker secret create B2_ACCOUNT_KEY -
```

If you do not feel comfortable copying secrets from your command line, you can use the wrapper `create_secret.sh`. This script prompts for a secret and ensures sensitive data is not displayed on your console. The script is available in the folder `/docker-secret` of your repository.

```console
local:~/restic-unattended$ ./docker-secret/create_secret.sh RESTIC_PASSWORD
local:~/restic-unattended$ ./docker-secret/create_secret.sh B2_ACCOUNT_ID
local:~/restic-unattended$ ./docker-secret/create_secret.sh B2_ACCOUNT_KEY
```

### Step 4 - Run Docker Service
Pending your choice to use environment variables or Docker secrets, you can deploy your service using Docker Compose or Stack Deploy.

#### Option 4a - Using Environment Variables
*Unchanged, however, use `dbm prod up -d` to run the production container in the background. Be aware that the production version of restic-unattended has no built-in shell support.*

#### Option 4b - Using Docker Secrets
Docker Swarm is needed to support external Docker secrets. As such, the services will be deployed as part of a Docker Stack in production. Deploy the stack using `docker-compose` as input. This ensures the environment variables are parsed correctly. The helper script `dbm` generates the configuration using the applicable `.yml` files and deploys the services to the stack `restic-unattended`. The provided `docker-compose.yml` does not invoke the schedule command yet, nor does it support a shell terminal. For testing purposes, deploy a development container `dev` first.

```console
local:~/restic-unattended$ dbm prod deploy
```

Run the following command to inspect the status of the Docker Stack.

```console
local:~/restic-unattended$ docker stack services restic-unattended
```

You should see the value `1/1` for `REPLICAS` for the restic service if the stack was initialized correctly. It might take a while before the services are up and running, so simply repeat the command after a few minutes if needed.

```console
ID  NAME                      MODE        REPLICAS  IMAGE                                      PORTS
*** restic-unattended_restic  replicated  1/1       markdumay/restic-unprivileged:0.5.0-debug   
```

Docker Swarm assigns a unique name to its deployed services. Retrieve the ID of the service by running `docker ps`.
```console
CONTAINER ID  IMAGE                                      COMMAND                 NAMES
ID            markdumay/restic-unprivileged:0.5.0-debug  "/bin/sh -c 'trap : …"  restic-unattended_restic.1.***
```

With the obtained ID, run the following command to validate the environment variables within the service are properly set.
```console
local:~/restic-unattended$ docker exec -it ID restic-unattended list
```

If the service is running as expected, remove the debug service as such.
```console
local:~/restic-unattended$ docker stack rm restic-unattended
```

Add a `schedule` command to the production image 'docker-compose.prod.yml` to ensure restic-unattended is running as daemon. See the <a href="#usage">next paragraph</a> for examples. Once done, deploy the production version of the image with the following command.
```console
local:~/restic-unattended$ dbm prod deploy
```

You can view the service logs with `docker service logs restic-unattended_restic` once the service is up and running. Debugging swarm services can be quite tedious. If for some reason your service does not initiate properly, you can get its task ID with `docker service ps restic-unattended_restic`. Running `docker inspect <task-id>` might give you some clues to what is happening. Use `docker stack rm restic-unattended` to remove the Docker stack entirely.

## Usage
*Restic-unattended* is intended to run from within a Docker container as an unattended service. As such, the most common use case is to define a schedule command in the `docker/docker-compose.yml` file. As an example, the below command instructs restic to perform a backup every 15 minutes and to remove obsolete snapshots daily at 01:00 am. It keeps the last 5 backups, the latest daily snapshot for the past 7 days, and the latest weekly snapshot for the last 13 weeks. In production, it is recommended to add the `--sustained` flag to ensure *restic-unattended* keeps running despite any errors. Be sure to use the `[""]` notation for the `cmd` specification, as the production-ready image has no built-in shell.

```yml
[...]
services:
    restic:
        [...]
        cmd: ["schedule", "0/15 * * * *", "-p=/data/backup", "--forget=0 1 * * *", "--keep-last=5", "--keep-daily=7", "--keep-weekly=13", "--sustained"]
```

*Restic-unattended* can also run from the command line (either from the host or from within a development container). The below list describes several scenarios.
* **From source (command line)** - the host machine requires Go 1.15 or later to be installed. From the `src` directory, run `go run main.go` followed by a specific command.
* **From source (Visual Studio Code)** - an `example-launch.json` file is included in the repository. Copy the configuration to `.vscode/launch.json` and set `env` and `args` as needed. The Go language tools need to be installed.
* **From within a container** - spin up a development container with `dbm dev up -t`. Run `restic-unattended` from within the container.

Several commands and flags are supported, which are described in the following paragraphs. They can also be inspected by using either `restic-unattended -h` or `restic-unattended <command> -h`.

### Backup
Creates a backup of the specified path and its subdirectories and stores it in a repository. The repository can be stored locally, or on a remote server. Backup connects to a previously initialized repository only, unless the flag `--init` is added.

```console
Usage:
  restic-unattended backup <path> [flags]

Flags:
  -h, --help          help for backup
  -H, --host string   hostname to use in backups (defaults to $HOSTNAME)
      --init          initialize the repository if it does not exist yet
```

### Check
The "check" command tests the repository for errors and reports any errors it finds. By default, the "check" command will always load all data directly from the repository and not use a local cache.

```console
Usage:
  restic-unattended check [flags]

Flags:
  -h, --help   help for check

Global Flags:
      --config string      config file (default is $HOME/.restic-unattended.yaml)
  -f, --logformat string   Log format to use: default, pretty, json (default "default")
  -l, --loglevel string    Level of logging to use: panic, fatal, error, warn, info, debug, trace (default "info")
```

### Forget
Forget removes old backups according to a rotation schedule. It both flags snapshots for removal as well as deletes (prunes) the actual old snapshot from the repository.

```console
Examples:
restic-unattended forget --keep-last 5
Keep the 5 most recent snapshots

restic-unattended forget --keep-daily 7
Keep the most recent backup for each of the last 7 days

Usage:
  restic-unattended forget [flags]

Flags:
      --keep-last int          never delete the n last (most recent) snapshots
      --keep-hourly int        for the last n hours in which a snapshot was made, keep only the last snapshot for each hour
      --keep-daily int         for the last n days which have one or more snapshots, only keep the last one for that day
      --keep-weekly int        for the last n weeks which have one or more snapshots, only keep the last one for that week
      --keep-monthly int       for the last n months which have one or more snapshots, only keep the last one for that month
      --keep-yearly int        for the last n years which have one or more snapshots, only keep the last one for that year
      --keep-tag stringArray   keep all snapshots which have all tags specified by this option (can be specified multiple times)
      --keep-within string     keep all snapshots which have been made within the duration of the latest snapshot
  -h, --help                   help for forget
```

### Help
Help provides help for any command in the application. Simply type `restic-unattended help [path to command]` for full details.

```console
Usage:
  restic-unattended help [command] [flags]

Flags:
  -h, --help   help for help
```

### List
*Restic-unattended* supports several environment variables on top of the default variables supported by restic. The additional variables typically end with a "_FILE" suffix. When initialized, *restic-unattended* reads the value from the specified variable file and maps it to the associated variable. This allows the initialization of Docker secrets as regular environment variables, restricted to the current process environment. Typically Docker secrets are mounted to the `/run/secrets` path, but this is not a prerequisite.

```console
Usage:
  restic-unattended list [flags]

Flags:
  -a, --all    Display all available variables
  -h, --help   help for list
```

### Restore
Restores a backup stored in a restic repository to a local path.

```console
Usage:
  restic-unattended restore <path> [flags]

Flags:
  -h, --help              help for restore
      --snapshot string   ID of the snapshot to restore (default "latest")
```

### Schedule
Schedule sets up a backup job that is repeated following a cron schedule. It optionally removes old snapshots using a policy too. The cron notation supports optional seconds. The following expressions are supported:

| Field name   | Mandatory? | Allowed values      | Allowed special characters |
| ----------   | ---------- | ------------------  | -------------------------- |
| Seconds      | No         | `0-59`              | `* / , -`                  |
| Minutes      | Yes        | `0-59`              | `* / , -`                  |
| Hours        | Yes        | `0-23`              | `* / , -`                  |
| Day of month | Yes        | `1-31`              | `* / , - ?`                |
| Month        | Yes        | `1-12` or `JAN-DEC` | `* / , -`                  |
| Day of week  | Yes        | `0-6` or `SUN-SAT`  | `* / , - ?`                |

Special characters:
* **Asterisk ( * )** - The asterisk indicates that the cron expression will match for all values of the field; e.g., using an asterisk in the 5th field (month) would indicate every month.

* **Slash ( / )** - Slashes are used to describe increments of ranges. For example 3-59/15 in the 1st field (minutes) would indicate the 3rd minute of the hour and every 15 minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...", that is, an increment over the largest possible range of the field. The form "N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the increment until the end of that specific range. It does not wrap around.

* **Comma ( , )** - 
Commas are used to separate items of a list. For example, using "MON,WED,FRI" in the 5th field (day of week) would mean Mondays, Wednesdays, and Fridays.

* **Hyphen ( - )** - 
Hyphens are used to define ranges. For example, 9-17 would indicate every hour between 9am and 5pm inclusive.

* **Question mark ( ? )** - 
Question mark may be used instead of '*' for leaving either day-of-month or day-of-week blank.

Predefined schedules:
The following predefined schedules may be used instead of the common cron fields:
`@yearly` (or `@annually`), `@monthly`, `@weekly`, `@daily` (or `@midnight`), and `@hourly`.

```console
Examples:
restic-unattended schedule '0 0,12 * * *'
Runs a scheduled backup at midnight and noon every day.

restic-unattended schedule '30 5 * * * *' --forget '0 0 * * *' --keep-daily 7
Runs a scheduled backup at minute 5 and 30 seconds of every hour. Recent 
backups are kept for each of the last 7 days, determined at 00:00 every day.

restic-unattended schedule '@weekly'
Runs a scheduled backup once a week at midnight on Sunday.

Usage:
  restic-unattended schedule <cron> [flags]

Flags:
      --forget string          remove old snapshots according to rotation schedule
  -h, --help                   help for schedule
  -H, --host string            hostname to use in backups (defaults to $HOSTNAME)
      --init                   initialize the repository if it does not exist yet
      --keep-daily int         for the last n days which have one or more snapshots, only keep the last one for that day
      --keep-hourly int        for the last n hours in which a snapshot was made, keep only the last snapshot for each hour
      --keep-last int          never delete the n last (most recent) snapshots
      --keep-monthly int       for the last n months which have one or more snapshots, only keep the last one for that month
      --keep-tag stringArray   keep all snapshots which have all tags specified by this option (can be specified multiple times)
      --keep-weekly int        for the last n weeks which have one or more snapshots, only keep the last one for that week
      --keep-within string     keep all snapshots which have been made within the duration of the latest snapshot
      --keep-yearly int        for the last n years which have one or more snapshots, only keep the last one for that year
  -p, --path string            local path to backup
      --sustained              sustain processing of scheduled jobs despite errors
```

### Snapshots
The "snapshots" command lists all snapshots stored in the repository.

```console
Usage:
  restic-unattended snapshots [flags]

Flags:
  -h, --help   help for snapshots
```

### Version
The "version" command displays information about the version of this software.

```console
Usage:
  restic-unattended version [flags]

Flags:
  -h, --help   help for version
```

### Global Flags
The following flags apply to all commands.
```console
Global Flags:
      --config string     config file (default is $HOME/.restic-unattended.yaml)
  -f, --logformat string  Log format to use: default, pretty, json (default "default")
  -l, --loglevel string   Level of logging to use: panic, fatal, error, warn, info, debug, trace (default "info")
```

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
<a href="https://github.com/markdumay/restic-unattended/blob/main/LICENSE" alt="License">
    <img src="https://img.shields.io/github/license/markdumay/restic-unattended.svg" />
</a>

Copyright © [Mark Dumay][blog]



<!-- MARKDOWN PUBLIC LINKS -->
[cobra_url]: https://github.com/spf13/cobra
[cron_url]: https://github.com/robfig/cron/
[cron_usage]: https://godoc.org/github.com/robfig/cron#hdr-Usage
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