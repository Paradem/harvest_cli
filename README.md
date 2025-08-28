# Harvest CLI

Simple CLI program written in GO to create a time entry for today, based on the
project and task that you select. Also you can provide the note on the CLI using
the `-n` flag.

This was also an experiment in using a coding agent to create much of the
application.

I found that while it could produce a lot of code quickly, I still had to
manually intervene and correct the errors. It made a lot of simple mistakes. But
overall I was able to use the code it provided. I even learned how to get it to
make edits. But some time it was a bit over zealous. 

Since I don't know the go programming language I would say that this process was
faster than if I had tried to write the code myself.

## TODO
Make better use of the bubbletea select options. Right now it is just using the
most basic values. None of the "huh?" widgets. It will just be an experiment to
try and get that working - right now what it has is more than enough for what
needs to be done.

## Build it

```bash
go build -o harvest_cli ./cmd
```

## Usage

You must set the environment variables for your api key.

The application uses:
HARVEST_ACCOUNT_ID
HARVEST_ACCESS_TOKEN
HARVEST_USER_ID (required for the `-e` flag)

Easiest way to do this is to use mise in your home dir. OR just set them in your
shell config - but since they need to be around for every project you should set
them somewhere global.

### Creating New Time Entries

On first run it will prompt you for the project and default task you want to
use in this directory. From then on you can just call it with the `-n` flag so
that you can set your message.

```bash
./harvest_cli -n "<your note goes here>"
```

### Selecting and Restarting Existing Time Entries

Use the `-e` flag to select from your existing time entries for today and restart them:

```bash
./harvest_cli -e
```

This will:
1. Show all your time entries for the current day
2. Allow you to select one using an interactive prompt
3. Restart the selected time entry (if it's not already running)

**Note**: The `HARVEST_USER_ID` environment variable must be set when using the `-e` flag.

I plan to incorporate this application into my `gh issues` work flow so that I
can choose the issue create the branch and start the timer all in one step.

Those scripts will end up in my dotfiles.
