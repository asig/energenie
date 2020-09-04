# energenie

This is a command line utility to control Energenie's programmable power 
strips EG-PM2-LAN, EG-PMS2-LAN, and EG-PMS2-WLAN.

It supports both EnerGenie's native protocol, as well as its HTTP protocol.

The power strips are totally awesome, but the management software is crap.
~~More importantly, there is no command line tool that runs under Linux.~~ 
There is a command line tool that also runs under Linux (`egctl`), but it
makes you write a config file first, which is ... um ... suboptimal. This
program fixes this.

# How to build
`go build`

# Usage
Usage: `energenie [flags] command [args]`

## Flags

| Flag         | Meaning                                                              |
|--------------|----------------------------------------------------------------------|
| `--address`  | The EnerGenie's network address. Default is 192.168.3.200            |
| `--port`     | The port to talk to if using the native protocol. Default is 5000    |
| `--password` | The password used to log in. Default is "1"                          |
| `--protocol` | The protocol to use, either 'native' or 'http'. Default is 'native'  |


## Supported commands

| Command                  | Meaning                                      |
|--------------------------|----------------------------------------------|
| `status [<socket-spec>]` | Print the sockets' status.                   |
| `on <socket-spec>`       | Turn sockets on that match `<socket-spec>`.  |
| `off <socket-spec>`      | Turn sockets off that match `<socket-spec>`. |

`<socket-spec>` is a comma-separated list of socket numbers, ranges, or 'all'
as a short cut for '1-4'.

Example: "1,2-3" specifies sockets 1, 2, and 3.
