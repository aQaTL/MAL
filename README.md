# Another MAL/AniList client

This time with CLI

- [Dependencies](#dependencies)
- [Quick start](#quick-start)
- [Commands](#commands) - usage of some commands
- [Examples](#examples) usage

## Dependencies

### For Linux

In order to have `mal copy` command working, you need to have either `xsel` or `xclip` installed.

## Quick start

If you have a working Go environment, you can download the app via `go get -u github.com/aqatl/mal`. 
Otherwise, download binaries from the [release](https://github.com/aQaTL/MAL/releases) page.

Remember that everything is stored in `$HOME/.mal` or `%userprofile%\.mal` (Windows).

### AniList mode

AniList mode is used by default. All you need to do to configure the app is to simply execute the program. 
It'll open AniList login page in your browser. Log in and authorize the app. And that's it - mal will cache 
the received token on your disk and use it to authenticate your requests. 

Run mal with `-r` flag to refresh cached lists.

### MyAnimeList mode

**Notice:** MAL API is shut down, currently it is not possible to access it.

To switch between AniList and MyAnimeList mode use the `s` command (e.g. `mal s`).

First, you need to give the app your credentials - username and password. To do that, execute 
`mal --prompt-credentials --verify --save-password`. If everything went good, you should see 
a list of 10 entries.

### Default behavior

The base command for everything is `mal`, which by default displays 10 last updated entries 
from your MAL. You can change the displayed list through some flags: 

```
--max value                    visible entries threshold (default: 0)
--a		               display all entries; same as max -1
--status value                 display entries only with given status [watching|planning|completed|repeating|paused|dropped]
--sort value                   display entries sorted by: [last-updated|title|episodes|score]
--reversed                     reversed list order
```

It's also good to run the app with `-r` (or `--refresh`) to update the cached list. Mind that there is not refresh interval so you have to refresh manually.

List of all commands and possible flags is available via `mal --help`. 

### Commands

All actions are done by variety of commands. They are in the following form:
`mal [global flags] command [command flags] [command arguments]`

Commands listed in `help` are divides into categories:

* **Update** command changes entry data and sends the updated version to your account
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
   mal sel [entry title]

CATEGORY:
   Config
```

For example, to select "Naruto", type `mal sel naruto` (case insensitive).
If `sel` is given no arguments, it will open a fuzzy search cui (console gui).

#### Update entry

For now, you can update your entry with the following commands: 

```
eps, episodes  Set the watched episodes value. If n not specified, the number will be increased by one
score          Set your rating for selected entry
status         Set your status for selected entry
cmpl           Alias for 'mal status completed'
delete, del    Delete entry
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
   mal status [watching|planning|completed|dropped|paused|repeating]

CATEGORY:
   Update
```

There is also `cmpl` command that is an alias for `status completed`.

Unfortunately, some commands may slightly differ between MyAnimeList and AniList mode and some may not
be present in both.

## Examples

A few examples of how I use this program.

Remember that everything is in `--help` :)

### Everyday usage

Okay, so when I add a new anime to my list, I run `mal -r` to update the cache. Then, if I 
want to watch it, I select it with `mal sel [name]`. Then I go to the web browser to find a 
website where I can watch it. If the name is long, I copy the title with `mal copy title`. 
To not forget the website and make it a little bit more convenient for me in the future, I 
copy the website's link and bind it to the selected anime with `mal web [website url]`. 

Now, when I want to watch it, I can just type `mal web` and it will open saved url in the 
web browser (you can configure which browser to use). When I finish an episode I type 
`mal eps` to update watched episodes and that's it. There's an option to automatically set 
the status to "completed", so I don't have to do anything more. 

Oh, and usually I also rate the show by `mal score [number from 0 to 10]`.

### Showing all entries from plan to watch list

Useful when you want to choose what to watch next.

`mal --status plantowatch --max -1`

The `--max -1` flag tells the program not to limit the displayed list length.

### Checking highest ranked (by you) shows

`mal --status all --sort score`

Again, you can add `--max -1` flag to turn off the list length limit.

### Showing your account stats

`mal stats`

### Checking entry details

`mal details`

`mal related`

`mal music`
