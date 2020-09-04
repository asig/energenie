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
	"net"
	"time"
)

type nativeClient struct {
	address  string
	password string
	port     int
	key      []byte
}

func NewNativeClient(address string, port int, password string) (Client, error) {
	if len(password) > 8 {
		return nil, fmt.Errorf("Password %q too long", password)
	}
	for len(password) < 8 {
		password = password + " "
	}
	key := make([]byte, 8)
	for i := 0; i < 8; i++ {
		key[i] = password[i]
	}

	return &nativeClient{address: address, port: port, key: key}, nil
}

func (c *nativeClient) connect() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.address, c.port), 4*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *nativeClient) solveTask(task []byte) []byte {
	res0 := uint16((task[0]^c.key[2]))*uint16(c.key[0]) ^ (uint16(c.key[6]) | (uint16(c.key[4]) << 8)) ^ uint16(task[2])
	res1 := uint16((task[1]^c.key[3]))*uint16(c.key[1]) ^ (uint16(c.key[7]) | (uint16(c.key[5]) << 8)) ^ uint16(task[3])

	res := []byte{byte(res0 & 0xff), byte(res0 >> 8), byte(res1 & 0xff), byte(res1 >> 8)}
	return res
}

func (c *nativeClient) Switch(sockets map[Socket]bool) error {
	_, err := c.statusAndSwitch(sockets)
	return err
}

func (c *nativeClient) Status() (states map[Socket]bool, names map[Socket]string, err error) {
	states, err = c.statusAndSwitch(map[Socket]bool{})
	if err != nil {
		return nil, nil, err
	}
	return states, nil, err
}

func (c *nativeClient) readState(conn net.Conn, task []byte) (map[Socket]bool, error) {
	res := make(map[Socket]bool)
	statcrypt := make([]byte, 4)
	_, err := conn.Read(statcrypt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 4; i++ {
		var on bool
		stat := (((statcrypt[i] - c.key[1]) ^ c.key[0]) - task[3]) ^ task[2]
		switch stat {
		case 0x22:
			on = false // protocol version 2.0
		case 0x11:
			on = true // protocol version 2.0
		case 0x41:
			on = true // protocol version 2.1
		case 0x82:
			on = false // protocol version 2.1
		case 0x51:
			on = true // WLAN version
		case 0x92:
			on = false // WLAN version
		}
		res[Socket(3-i+1)] = on
	}
	return res, nil
}

func (c *nativeClient) statusAndSwitch(newState map[Socket]bool) (map[Socket]bool, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res := make(map[Socket]bool)

	_, err = conn.Write([]byte{0x11})
	if err != nil {
		return nil, err
	}

	task := make([]byte, 4)
	_, err = conn.Read(task)
	if err != nil {
		return nil, err
	}

	solution := c.solveTask(task)
	_, err = conn.Write(solution)
	if err != nil {
		return nil, err
	}

	_, err = c.readState(conn, task)
	if err != nil {
		return nil, err
	}

	ctrlcrypt := make([]byte, 4)
	for i := 0; i < 4; i++ {
		val := byte(0x04) // default to "not switched"
		if on, ok := newState[Socket(3-i+1)]; ok {
			val = 0x2
			if on {
				val = 0x1
			}
		}
		ctrlcrypt[i] = (((val ^ task[2]) + task[3]) ^ c.key[0]) + c.key[1]
	}
	_, err = conn.Write(ctrlcrypt)
	if err != nil {
		return nil, err
	}

	res, err = c.readState(conn, task)
	if err != nil {
		return nil, err
	}


	// Do we need to send dummy schedule? Right now, the device does not react for 10 secs or so after a successful
	// execution...

	return res, nil
}
