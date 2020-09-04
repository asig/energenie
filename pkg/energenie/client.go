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

type Socket int

const (
	Socket_Min = Socket(1)
	Socket_Max = Socket(4)
)

type Client interface {
	 Switch(sockets map[Socket]bool) error
	 Status() (states map[Socket]bool, names map[Socket]string, err error)
}
