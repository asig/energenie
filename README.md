# energenie

This is a command line utility to control Energenie's programmable power 
strips EG-PM2-LAN, EG-PMS2-LAN, and EG-PMS2-WLAN.

The power strips are totally awesome, but the management software is crap.
More importantly, there is no command line tool that runs under Linux. This
program fixes this.

# How to build
`go build`

# Usage
Usage: `energenie [flags] command [args]`

Flags:
- `--address`: The EnerGenie's network address. Default is 192.168.3.200
- `--password`: The password used to log in. Default is "1"

Supported commands:
- `status` [<socket-spec>]: Print the sockets' status.
- `on` <socket-spec>: Turn sockets on that match <socket-spec>.
- `off` <socket-spec>: Turn sockets off that match <socket-spec>.

<socket-spec> can be a comma-separated list of socket numbers (1 - 4),
ranges, or 'all' as a short cut for 1-4.append

Example: "1,2-3" specifies sockets 1, 2, and 3.
