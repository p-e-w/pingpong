// PingPong: End-to-end latency measurement for Matrix
// Copyright (C) 2021  Philipp Emanuel Weidmann <pew@worldwidemann.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type client struct {
	*mautrix.Client
	homeserver string
	username   string
	password   string
	userID     string
}

func (c *client) UnmarshalText(text []byte) (err error) {
	parts := strings.SplitN(string(text), ":", 3)

	if len(parts) != 3 || !strings.HasPrefix(parts[0], "@") {
		return errors.New("user credentials must be of the form @user:homeserver.org:password")
	}

	c.homeserver = parts[1]
	c.username = strings.TrimPrefix(parts[0], "@")
	c.password = parts[2]
	c.userID = fmt.Sprintf("@%v:%v", c.username, c.homeserver)

	c.Client, err = mautrix.NewClient(c.homeserver, "", "")

	return
}

func (c *client) login() (err error) {
	_, err = c.Login(&mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: c.username,
		},
		Password:         c.password,
		StoreCredentials: true,
	})

	return
}

func (c *client) onMessage(senderID id.UserID, roomID id.RoomID, callback func(*event.Event)) (err error) {
	syncer := c.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Sender == senderID && evt.RoomID == roomID {
			callback(evt)
		}
	})

	err = c.Sync()

	return
}

func (c *client) formatLogEntry(format string, v ...interface{}) string {
	return fmt.Sprintf("[%v] %v", c.userID, fmt.Sprintf(format, v...))
}

func (c *client) log(format string, v ...interface{}) {
	if cli.Debug {
		log.Print(c.formatLogEntry(format, v...))
	}
}

func (c *client) error(format string, v ...interface{}) {
	c.log("[ERROR] "+format, v...)
}

func (c *client) fatal(format string, v ...interface{}) {
	panic(errors.New(c.formatLogEntry("[FATAL] "+format, v...)))
}
