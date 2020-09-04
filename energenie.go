/*
 * Copyright (c) 2018, 2020 Andreas Signer <asigner@gmail.com>
 *
 * This file is part of energenie.
 *
 * energenie is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * energenie is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/asig/energenie/pkg/energenie"
)

var (
	client energenie.Client

	flagAddress  = flag.String("address", "192.168.3.200", "EnerGenie's address")
	flagPort = flag.Int("port", 5000, "Port used for native protocol")
	flagPassword = flag.String("password", "1", "password to log in to EnerGenie")
	flagProtocol = flag.String("protocol", "native", "protocol to use. Possible values are 'native' and 'http'")
)

func init() {
	var err error
	flag.Parse()
	switch strings.ToLower(*flagProtocol) {
	case "native":
		client, err = energenie.NewNativeClient(*flagAddress, *flagPort, *flagPassword)
	case "http":
		client, err = energenie.NewHttpClient(*flagAddress, *flagPassword)
	default: log.Fatalf("Unsupported protocol %q. Valid values are \"native\" and \"http\"", *flagProtocol);
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't instantiate a client: %s", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: energenie [flags] command [args]

flags:
  --address:  The EnerGenie's network address
  --port:     The port used in the native protocol
  --password: The password used to log in
  --protocol: What protocol to use: "http" or "native"
  
commands:
  status [<socket-spec>]: Print the sockets' status.
  on <socket-spec>:       Turn sockets on that match <socket-spec>.
  off <socket-spec>:      Turn sockets off that match <socket-spec>.

  <socket-spec> can be a comma-separated list of socket numbers (1 - 4),
  ranges, or 'all' as a short cut for 1-4.

  Example: 1,2-3 specifies sockets 1, 2, and 3.
`)
	os.Exit(1)
}

func atoiOrDie(raw string) int {
	i, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s is not a valid integer\n", raw)
		usage()
	}
	return i
}

func addAll(set map[energenie.Socket]bool) map[energenie.Socket]bool {
	for i := energenie.Socket_Min; i <= energenie.Socket_Max; i++ {
		set[i] = true
	}
	return set
}

func checkSocket(socket energenie.Socket) {
	if socket < energenie.Socket_Min || socket > energenie.Socket_Max {
		fmt.Fprintf(os.Stderr, "Invalid socket number %d, must be between %d and %d.\n", socket, energenie.Socket_Min, energenie.Socket_Max)
		os.Exit(1)
	}
}

func addRange(r string, set map[energenie.Socket]bool) map[energenie.Socket]bool {
	parts := strings.Split(r, "-")
	low := energenie.Socket(atoiOrDie(parts[0]))
	checkSocket(low)
	hi := energenie.Socket(atoiOrDie(parts[1]))
	checkSocket(hi)
	if low > hi {
		low, hi = hi, low
	}
	for i := low; i <= hi; i++ {
		set[i] = true
	}
	return set
}

func addSingle(str string, set map[energenie.Socket]bool) map[energenie.Socket]bool {
	s := energenie.Socket(atoiOrDie(strings.TrimSpace(str)))
	checkSocket(s)
	set[s] = true
	return set
}


type socketSlice []energenie.Socket
func (p socketSlice) Len() int           { return len(p) }
func (p socketSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p socketSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func parseSocketSpec(specString string) []energenie.Socket {
	seen := make(map[energenie.Socket]bool)
	parts := strings.Split(specString, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "all" {
			seen = addAll(seen)
		} else if strings.Index(part, "-") > 0 {
			seen = addRange(part, seen)
		} else {
			seen = addSingle(part, seen)
		}
	}

	var res []energenie.Socket
	for key := range seen {
		res = append(res, key)
	}
	sort.Sort(socketSlice(res))
	return res
}

func showStatus(socketSpec string) {
	sockets := parseSocketSpec(socketSpec)

	state, names, err := client.Status()
	hasNames := names != nil
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't get status: %s", err)
		os.Exit(1)
	}
	for _, socket := range sockets {
		status := "off"
		if state[socket] {
			status = "on"
		}
		socketName := fmt.Sprintf("%d", socket)
		if hasNames {
			socketName = socketName + " (" + names[socket] + ")"
		}
		fmt.Printf("Socket %s: %s\n", socketName, status)
	}
}

func switchSockets(on bool, socketSpec string) {
	sockets := parseSocketSpec(socketSpec)
	states := make(map[energenie.Socket]bool)
	for _, s := range sockets {
		states[s] = on
	}
	err := client.Switch(states)
	if err != nil {
			fmt.Fprintf(os.Stderr, "Can't switch sockets: %s", err)
			os.Exit(1)
	}
}

func main() {
	args := flag.Args()
	if len(args) == 0 || len(args) > 2 {
		usage()
		os.Exit(1)
	}

	switch args[0] {
	case "status":
		if len(args) > 2 {
			usage()
		}
		s := "all"
		if len(args) == 2 {
			s = args[1]
		}
		showStatus(s)
	case "on", "off":
		if len(args) != 2 {
			usage()
		}
		switchSockets(args[0] == "on", args[1])
	default:
		fmt.Fprintf(os.Stderr, "%s is not a valid command.\n", args[0])
		usage()
	}
}
