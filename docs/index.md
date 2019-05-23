# system-gen

Generates systemd files in npm-like style. 

Groups several units in one, so administrator can start/stop all services by one command.
No special scripts (except install helpers) or commands - all 'vanila' systemd units. 

Supports:
* services
* one-shots
* timers

# Installation

* Use releases page on github

or

* Build from sources (needs go toolchain) `go get -v -u github.com/reddec/system-gen/cmd/...` 

## Example

Assume you have some synchronization software that should be started automatically every hour
plus system administrator should  have option to start process manually keeping all settings 'as is'.

So let's imagine there are those files:

* `my-sync-script.sh` - script to synchronize something. Requires environment `TOKEN`
* ..that's all =)

### Init project

Let's create new project and call it project **my-sync**.

```
system-gen init my-sync
```

This command creates new directory `my-sync` with all required content. 
Change dir to the directory.


### Add command to start

We need wrap our script as one-shot (because when job done script should stop) unit called 
 `sync` with all required environment variables (token for our case)

```
system-gen add oneshot -e TOKEN:MYTOKEN sync path/to/my-sync-script.sh
```

This command creates a one-shot unit service with environment variables.

hint: look ar file system.json


### Add timer

We want to start our service every hour so we need to create timer that should launch our
one-short service `sync`

```
system-gen add timer sync 1h
```


### Generate everything


Now let our generator do all heavy job and generates all systemd files and helpers

```
system-gen generate
```


Look at folder `generated`:

* `my-sync.service` - group unit file. Start/Stop this service to start/stop everything
* `my-sync-sync.service` - wrapper of our script that runs, waits and exits
* `my-sync-timer-1h.timer` - timer that starts script every hour
* `install.sh` - install all services
* `uninstall.sh` - uninstall all services from the project



Run `install.sh` and then `systemctl start my-sync` to start project!