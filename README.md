## PingPong: End-to-end latency monitoring for Matrix

![build](https://github.com/p-e-w/pingpong/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/p-e-w/pingpong)](https://goreportcard.com/report/github.com/p-e-w/pingpong)
![GPL-3.0 License](https://img.shields.io/github/license/p-e-w/pingpong)
[![#pingpong:matrix.org](https://img.shields.io/matrix/pingpong:matrix.org?label=%23pingpong%3Amatrix.org)](https://matrix.to/#/#pingpong:matrix.org)

PingPong measures transport latencies on [Matrix](https://matrix.org/) networks.
It connects to two Matrix accounts simultaneously, and bounces messages
back and forth between them. It aggregates all information in an intuitive
terminal user interface, and automatically computes statistics.

![Screenshot](https://user-images.githubusercontent.com/2702526/104276628-481ebb80-54cb-11eb-8b13-dc53378a6c08.png)

PingPong aims to be a good citizen on homeservers. It gracefully recovers
from various temporary error conditions, has built-in rate limiting mechanisms,
and cleans up after itself to the maximum extent possible in the Matrix
ecosystem, even if an error occurs or the program crashes.

PingPong is built using the excellent
[Mautrix](https://github.com/tulir/mautrix-go),
[Tcell](https://github.com/gdamore/tcell),
and [Kong](https://github.com/alecthomas/kong) packages.


## Installation and usage

PingPong requires Go (Golang) to build. The build process is simple:

```
git clone https://github.com/p-e-w/pingpong.git
cd pingpong
go run . --help
```

This prints a help message explaining the command line interface,
which is all you need to know in order to start using PingPong.


## What it measures

The **total time** metric represents the time from a message being
prepared and sent by the sender to the message having been received,
parsed, and processed by the recipient. Preparation, parsing, and
processing should take on the order of microseconds in most cases,
so this is the end-to-end transport latency for practical purposes.
This time is calculated using a single local monotonic clock and is
therefore *always* accurate.

The **client -> server** metric is the time from a message being
prepared and sent to the message being forwarded by the sender's
homeserver to the recipient's homeserver, as indicated by the event's
`origin_server_ts` field. Note that the send time is measured by
the local clock whereas the `origin_server_ts` field is populated
using the homeserver's clock. Therefore, the accuracy of this metric
depends on how closely those two clocks are synchronized.

The **server -> server** metric is the time from a message being
forwarded by the sender's homeserver to the message being processed
by the recipient's homeserver, as indicated by the event's
`unsigned.age` field. The accuracy of this metric depends on how
closely the two homeservers' clocks are synchronized.

The **server -> client** metric is the time from a message being
processed by the recipient's homeserver to the message having been
received, parsed, and processed by the recipient, as indicated by
a combination of the event's `origin_server_ts` and `unsigned.age`
fields. The accuracy of this metric depends on how closely the
local clock and the two homeservers' clocks are synchronized.

PingPong performs basic sanity checks on those metrics, and if they
indicate that any clocks are out of sync or the event's timestamps are
otherwise unreliable, breakdown metrics are not displayed.


## Similar projects

[echo](https://github.com/maubot/echo) is a Matrix bot that replies with
timing information when sent a ping message. It is used to generate the
ping ranking in the "This Week in Matrix" blog posts.

[matrix-monitor-bot](https://github.com/turt2live/matrix-monitor-bot) is
another Matrix bot that measures message latency, and serves the results
as Prometheus metrics, suitable for integration into an existing monitoring
solution.

These bots work, but I wanted a traditional command line tool with
everything included, zero configuration required, and a beautiful
terminal UI. That is what PingPong tries to be.


## License

Copyright &copy; 2021  Philipp Emanuel Weidmann (<pew@worldwidemann.com>)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

**By contributing to this project, you agree to release your
contributions under the same license.**
