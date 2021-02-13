# SSH Utils

## Features
 - **ls** - list active sessions
 - **attach** - attach (view and control) a remote ssh session
 - **kill** - kill an ssh session
 - **log**(coming soon) - begin to log outputs of all ssh sessions to log files
 - **watch**(coming soon) - notifies you about new ssh sessions

## Installation

### From source

```bash
GO111MODULE=auto go get github.com/td0m/sshutils
sudo sshutils
```

Please make sure `$GOPATH` is set and `$GOPATH/bin` is in your `$PATH`.

### AUR

```
yay -S sshutils
sudo sshutils
```
