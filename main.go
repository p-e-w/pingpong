// PingPong: End-to-end latency monitoring for Matrix
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
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var cli struct {
	UserOne       client        `kong:"arg,name='user1',help='Credentials for the first user, of the form @user:homeserver.org:password.'"`
	UserTwo       client        `kong:"arg,name='user2',help='Credentials for the second user, of the form @user:homeserver.org:password.'"`
	MessageText   string        `kong:"short='m',default='ping',help='Content of the messages sent back and forth.'"`
	Interval      time.Duration `kong:"short='i',default='3s',help='Time to wait before responding to a message.'"`
	RetryInterval time.Duration `kong:"short='r',default='5s',help='Time to wait before retrying an operation if an error occurs.'"`
	Debug         bool          `kong:"short='d',help='Print a detailed log of every operation, instead of the default TUI.'"`
}

var one client
var two client

type latency struct {
	total        time.Duration
	clientServer time.Duration
	serverServer time.Duration
	serverClient time.Duration
}

var oneLatencies chan latency
var twoLatencies chan latency

var latencyBreakdownValid = true

var roomID id.RoomID

var sendTimes = make(map[id.EventID]time.Time)

func main() {
	kong.Parse(&cli)

	one = cli.UserOne
	two = cli.UserTwo

	quit := make(chan os.Signal, 1)

	if cli.Debug {
		signal.Notify(quit, os.Interrupt)
	} else {
		oneLatencies = make(chan latency)
		twoLatencies = make(chan latency)

		err := initTUI()
		if err != nil {
			log.Fatalf("unable to initialize terminal UI: %v", err)
		}

		go func() {
			runTUI()
			close(quit)
		}()
	}

	defer func() {
		if err := recover(); err != nil {
			if !cli.Debug {
				quitTUI()
				// Wait for runTUI() to return to ensure the error
				// is not printed to the alternate screen
				<-quit
			}
			log.Fatal(err)
		}
	}()

	login(&one)
	defer logout(&one)
	login(&two)
	defer logout(&two)

	resp, err := one.CreateRoom(&mautrix.ReqCreateRoom{
		Preset: "public_chat",
	})
	if err != nil {
		one.fatal("unable to create room: %v", err)
	}
	roomID = resp.RoomID
	one.log("created room %v", roomID)
	defer leave(&one)

	_, err = two.JoinRoom(string(roomID), "", nil)
	if err != nil {
		two.fatal("unable to join room %v: %v", roomID, err)
	}
	two.log("joined room %v", roomID)
	defer leave(&two)

	go func() {
		oneMessages := make(chan *event.Event)
		twoMessages := make(chan *event.Event)

		go listen(&one, &two, oneMessages)
		go listen(&two, &one, twoMessages)

		ping(&one)

		for {
			select {
			case evt := <-oneMessages:
				process(&one, evt, oneLatencies)
			case evt := <-twoMessages:
				process(&two, evt, twoLatencies)
			}
		}
	}()

	<-quit
}

func login(c *client) {
	err := c.login()
	if err != nil {
		c.fatal("unable to log in: %v", err)
	}
	c.log("logged in")
}

func logout(c *client) {
	_, err := c.Logout()
	if err != nil {
		c.fatal("unable to log out: %v", err)
	}
	c.log("logged out")
}

func leave(c *client) {
	_, err := c.LeaveRoom(roomID)
	if err != nil {
		c.fatal("unable to leave room %v: %v", roomID, err)
	}
	c.log("left room %v", roomID)
}

func listen(receiver *client, sender *client, messages chan *event.Event) {
	for {
		err := receiver.onMessage(sender.UserID, roomID, func(evt *event.Event) {
			messages <- evt
		})
		if err != nil {
			receiver.error("unable to sync: %v", err)
		}
		time.Sleep(cli.RetryInterval)
	}
}

func ping(c *client) {
	for {
		sendTime := time.Now()
		resp, err := c.SendText(roomID, cli.MessageText)
		if err != nil {
			c.error("unable to send message: %v", err)
			time.Sleep(cli.RetryInterval)
			continue
		}
		c.log("sent message %v", resp.EventID)
		sendTimes[resp.EventID] = sendTime
		break
	}
}

func process(c *client, evt *event.Event, latencies chan latency) {
	receiveTime := time.Now()

	sendTime, ok := sendTimes[evt.ID]
	if !ok {
		return
	}

	delete(sendTimes, evt.ID)

	totalLatency := receiveTime.Sub(sendTime)
	originServerClientLatency := receiveTime.Sub(time.Unix(0, evt.Timestamp*int64(time.Millisecond)))
	serverServerLatency := time.Duration(evt.Unsigned.Age) * time.Millisecond

	latency := latency{
		total:        totalLatency,
		clientServer: totalLatency - originServerClientLatency,
		serverServer: serverServerLatency,
		serverClient: originServerClientLatency - serverServerLatency,
	}

	if latency.clientServer <= 0 || latency.serverServer <= 0 || latency.serverClient <= 0 {
		// This is a permanent flag and applies to both transmission directions,
		// because any such occurrence casts doubt on the synchronization of the entire system,
		// even if later data or data from the other direction appears to be valid.
		latencyBreakdownValid = false
	}

	breakdownValidityMessage := ""
	if !latencyBreakdownValid {
		breakdownValidityMessage = " (BREAKDOWN INVALID)"
	}

	c.log(
		"received message %v, time: %v total, %v client->server, %v server->server, %v server->client%v",
		evt.ID,
		latency.total.Round(time.Millisecond),
		latency.clientServer.Round(time.Millisecond),
		latency.serverServer.Round(time.Millisecond),
		latency.serverClient.Round(time.Millisecond),
		breakdownValidityMessage,
	)

	if latencies != nil {
		latencies <- latency
	}

	time.Sleep(cli.Interval)

	ping(c)
}
