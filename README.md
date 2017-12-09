# Another MAL client

This time with CLI

[![asciicast](https://asciinema.org/a/DNvVHEadubTfNeZo6O84SVuaO.png)](https://asciinema.org/a/DNvVHEadubTfNeZo6O84SVuaO)

## Dependencies

### For Linux

In order to have `mal copy` command working, you need to have either `xsel` or `xclip` installed.

## Quick start

If you have a working Go environment, you can download the app via `go get -u github.com/aqatl/mal`. 
Otherwise, download binaries from the [release](https://github.com/aQaTL/MAL/releases) page.

First, you need to give the app your credentials - username and password (everything is 
stored in `$HOME/.mal` or `%userprofile%\.mal` (Windows)). To do that, execute 
`mal --prompt-credentials --verify --save-password`. If everything went good, you should see 
a list of 10 entries.

The base command for everything is `mal`, which by default displays 10 last updated entries 
from your MAL. You can change the displayed list through some flags: 

```
--max value                    visible entries threshold (default: 0)
--status value                 display entries only with given status [watching|completed|onhold|dropped|plantowatch]
--sort value                   display entries sorted by: [last-updated|title|episodes|score]
--reversed                     reversed list order
```

It's also good to run the app with `-r` (or `--refresh`) to update the cached list.

List of all commands and possible flags is available via `mal --help`. 

### Commands

All actions are done by variety of commands. They are in the following form:
`mal [global flags] command [command flags] [command arguments]`

Commands listed in `help` are divides into categories:

* **Update** command changes entry data and sends the updated version to MAL service
* **Action** command performs action that uses the entry data like printing it to the console
* **Config** command manipulates on the app configuration file (look at `mal cfg --help` for details)

You can always see the details of the specific command via `help` like this: 
`mal <command> --help`

#### Select entry to work with

Commands that use entry data need to know which entry you want to use. And there's a thing 
called "selected entry". To select an entry, use the `mal sel` command. And here's a usage 
of that command (` mal sel --help`): 

```
NAME:
   mal sel - Select an entry

USAGE:
   mal sel [entry ID]

CATEGORY:
   Config

OPTIONS:
   -t  Select entry by title instead of by ID
```

For example, to select "Naruto", type `mal sel -t naruto` (case insensitive)

#### Update entry

For now, you can update your entry with the following commands: 

```
eps, episodes  Set the watched episodes value. If n not specified, the number will be increased by one
score          Set your rating for selected entry
status         Set your status for selected entry
```

##### `mal eps` command

```
NAME:
   mal eps - Set the watched episodes value. If n not specified, the number will be increased by one

USAGE:
   mal eps <n>

CATEGORY:
   Update
```

There's an option to have mal automatically turn the entry status to completed after updating 
the watched episodes value. To do that, use the `status-auto-update` config command. 

```
NAME:
   mal cfg status-auto-update - Allows entry to be automatically set to completed when number of all episodes is reached or exceeded

USAGE:
   mal cfg status-auto-update [off|normal|after-threshold]
```

As you can see, there are 2 modes of auto-update: normal and after-threshold.

The first behaves as you would expect -> the status is changes when entry has 12 episodes 
and you hit the 12 watched episodes.

As for the `after-threshold`, the status will change after you exceed the number of 
episodes. For example: when entry has 12 episodes and you hit 13 -> status is changed to 
completed and your watched entries value is changed back to 12.

##### `mal score` command

```
NAME:
   mal score - Set your rating for selected entry

USAGE:
   mal score <0-10>

CATEGORY:
   Update
```

##### `mal status` command

```
NAME:
   mal status - Set your status for selected entry

USAGE:
   mal status [watching|completed|onhold|dropped|plantowatch]

CATEGORY:
   Update
```


