## PingPong: End-to-end latency measurement for Matrix

![build](https://github.com/p-e-w/pingpong/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/p-e-w/pingpong)](https://goreportcard.com/report/github.com/p-e-w/pingpong)
![GPL-3.0 License](https://img.shields.io/github/license/p-e-w/pingpong)
[![#pingpong:matrix.org](https://img.shields.io/matrix/pingpong:matrix.org?label=%23pingpong%3Amatrix.org)](https://matrix.to/#/#pingpong:matrix.org)

PingPong measures transport latencies on [Matrix](https://matrix.org/) networks.
It connects to two Matrix accounts simultaneously, and bounces messages
back and forth between them. It aggregates all information in an intuitive
terminal user interface, and automatically computes statistics.

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
