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

package energenie

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type httpClient struct {
	client *http.Client
	address string
	password string
}

func NewHttpClient(address, password string) (Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	return &httpClient{client: client, address: address, password: password}, nil
}

func (c *httpClient) Switch(sockets map[Socket]bool) error {
	c.login()
	defer c.logout()
	for s, state := range sockets {
		err := c.swtch(s, state)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *httpClient) swtch(s Socket, on bool) error {
	val := "0"
	if on {
		val = "1"
	}
	form := make(map[string][]string)
	form[fmt.Sprintf("cte%d", s)] = []string{val}
	resp, err := c.client.PostForm(fmt.Sprintf("http://%s/", c.address), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func (c *httpClient) Status() (states map[Socket]bool, names map[Socket]string, err error) {
	body := c.login()
	defer c.logout()

	rawNames := extractNames(body)
	i1 := strings.Index(body, "sockstates = ")
	body = body[i1:]
	i1 = strings.Index(body, "[")
	i2 := strings.Index(body, "]")
	rawStates := strings.Split(body[i1+1:i2], ",")

	states = make(map[Socket]bool)
	names =  make(map[Socket]string)
	for s := Socket_Min; s <= Socket_Max; s++ {
		if rawStates[s-1] == "1" {
			states[s] = true
		}
		names[s] = rawNames[s-1]
	}
	return states, names, nil
}

func (c *httpClient) logout() string {
	resp, _ := c.client.Get(fmt.Sprintf("http://%s/login.html", c.address))
	body, _ := ioutil.ReadAll(resp.Body)
	res := string(body)
	resp.Body.Close()
	return res
}

func (c *httpClient) login() string {
	resp, _ := c.client.PostForm(fmt.Sprintf("http://%s/login.html", c.address), map[string][]string{
		"pw": []string{c.password},
	})
	body, _ := ioutil.ReadAll(resp.Body)
	res := string(body)
	resp.Body.Close()
	return res
}

func extractNames(html string) []string {
	var names []string
	pos := 0
	maxNameLen := 0
	var n *string
	for i := Socket_Min; i <= Socket_Max; i++ {
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

func indexAfter(haystack, needle string, pos int) int {
	s := haystack[pos:]
	i := strings.Index(s, needle)
	if i > -1 {
		i += pos
	}
	return i
}


