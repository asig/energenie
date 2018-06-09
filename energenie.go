/*
 * energenie: A command line utilitly to control Energenie's programmable
 * power strips.
 *
 * Copyright (c) 2018 Andreas Signer <asigner@gmail.com>
 *
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
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
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	client *http.Client

	flagAddress  = flag.String("address", "192.168.3.200", "EnerGenie's address")
	flagPassword = flag.String("password", "1", "password to log in to EnerGenie")
)

const (
	minSocket = 1
	maxSocket = 4
)

type socketList []int

func init() {
	flag.Parse()

	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: energenie [flags] command [args]

flags:
  --address:  The EnerGenie's network address
  --password: The password used to log in
  
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

func addAll(set map[int]bool) map[int]bool {
	for i := minSocket; i <= maxSocket; i++ {
		set[i] = true
	}
	return set
}

func checkSocket(socket int) {
	if socket < minSocket || socket > maxSocket {
		fmt.Fprintf(os.Stderr, "Invalid socket number %d, must be between %d and %d.\n", socket, minSocket, maxSocket)
		os.Exit(1)
	}
}

func addRange(r string, set map[int]bool) map[int]bool {
	parts := strings.Split(r, "-")
	low := atoiOrDie(parts[0])
	checkSocket(low)
	hi := atoiOrDie(parts[1])
	checkSocket(hi)
	if low > hi {
		low, hi = hi, low
	}
	for i := low; i <= hi; i++ {
		set[i] = true
	}
	return set
}

func addSingle(str string, set map[int]bool) map[int]bool {
	s := atoiOrDie(strings.TrimSpace(str))
	checkSocket(s)
	set[s] = true
	return set
}

func parseSocketSpec(specString string) socketList {
	seen := make(map[int]bool)
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

	var res socketList
	for key := range seen {
		res = append(res, key)
	}
	sort.Ints(res)
	return res
}

func logout() string {
	resp, _ := client.Get(fmt.Sprintf("http://%s/login.html", *flagAddress))
	body, _ := ioutil.ReadAll(resp.Body)
	res := string(body)
	resp.Body.Close()
	return res
}

func login() string {
	resp, _ := client.PostForm(fmt.Sprintf("http://%s/login.html", *flagAddress), map[string][]string{
		"pw": []string{*flagPassword},
	})
	body, _ := ioutil.ReadAll(resp.Body)
	res := string(body)
	resp.Body.Close()
	return res
}

func indexAfter(haystack, needle string, pos int) int {
	s := haystack[pos:]
	i := strings.Index(s, needle)
	if i > -1 {
		i += pos
	}
	return i
}

func extractName(html string, firstPos int) (res *string, pos int) {
	beginMarker := "<h2 class=\"ener\">"
	endMarker := "</h2>"
	begin := indexAfter(html, beginMarker, firstPos)
	if begin == -1 {
		return nil, -1
	}
	begin += len(beginMarker)
	end := indexAfter(html, endMarker, begin)
	name := html[begin:end]
	return &name, begin
}

func extractNames(html string) []string {
	var names []string
	pos := 0
	maxNameLen := 0
	var n *string
	for i := minSocket; i <= maxSocket; i++ {
		n, pos = extractName(html, pos)
		if n == nil {
			break
		}
		name := strings.TrimSpace(*n)
		names = append(names, name)
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	for idx, _ := range names {
		for len(names[idx]) < maxNameLen {
			names[idx] = names[idx] + " "
		}
	}

	return names
}

func showStatus(socketSpec string) {
	sockets := parseSocketSpec(socketSpec)

	body := login()
	defer logout()

	names := extractNames(body)
	i1 := strings.Index(body, "sockstates = ")
	body = body[i1:]
	i1 = strings.Index(body, "[")
	i2 := strings.Index(body, "]")
	states := strings.Split(body[i1+1:i2], ",")

	for idx, socket := range sockets {
		status := "off"
		if states[socket-1] == "1" {
			status = "on"
		}
		fmt.Printf("Socket %d (%s): %s\n", socket, names[idx], status)
	}
}

func switchSocket(on bool, socketSpec string) {
	sockets := parseSocketSpec(socketSpec)

	val := "0"
	if on {
		val = "1"
	}

	login()
	defer logout()

	for _, socket := range sockets {
		form := make(map[string][]string)
		form["cte1"] = []string{""}
		form["cte2"] = []string{""}
		form["cte3"] = []string{""}
		form["cte4"] = []string{""}
		form[fmt.Sprintf("cte%d", socket)] = []string{val}

		resp, _ := client.PostForm(fmt.Sprintf("http://%s/", *flagAddress), form)
		resp.Body.Close()
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
		switchSocket(args[0] == "on", args[1])
	default:
		fmt.Fprintf(os.Stderr, "%s is not a valid command.\n", args[0])
		usage()
	}
}
