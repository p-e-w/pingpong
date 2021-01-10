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
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var barCharacters = []rune{' ', '\u2581', '\u2582', '\u2583', '\u2584', '\u2585', '\u2586', '\u2587', '\u2588'}

var placeholder = strings.Repeat("\u00B7", 5)

var (
	frameStyle       = tcell.StyleDefault.Foreground(tcell.Color250).Background(tcell.Color234)
	frameBrightStyle = frameStyle.Foreground(tcell.Color231)
	oneUserIDStyle   = frameStyle.Foreground(tcell.Color197)
	twoUserIDStyle   = frameStyle.Foreground(tcell.Color106)
	graphStyle       = tcell.StyleDefault.Foreground(tcell.Color240).Background(tcell.Color16)
	graphBrightStyle = graphStyle.Foreground(tcell.Color250)
	oneBarStyle      = graphStyle.Foreground(tcell.Color64)
	twoBarStyle      = graphStyle.Foreground(tcell.Color161)
)

type st struct {
	style tcell.Style
	text  string
}

var screen tcell.Screen

const latencyHistorySize = 500

type latencyData struct {
	list  []latency
	sum   latency
	count int
}

var oneData latencyData
var twoData latencyData

type bar struct {
	styleOne bool
	latency  latency
}

func initTUI() (err error) {
	screen, err = tcell.NewScreen()
	if err != nil {
		return
	}

	err = screen.Init()

	return
}

func runTUI() {
	draw()

	events := make(chan tcell.Event)

	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	for {
		select {
		case latency := <-oneLatencies:
			update(&oneData, latency)
		case latency := <-twoLatencies:
			update(&twoData, latency)
		case evt := <-events:
			switch evt := evt.(type) {
			case *tcell.EventResize:
				screen.Sync()
				draw()
			case *tcell.EventKey:
				if evt.Key() == tcell.KeyEscape {
					screen.Fini()
					return
				}
			}
		}
	}
}

func quitTUI() {
	screen.PostEventWait(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
}

func update(d *latencyData, latency latency) {
	d.list = append(d.list, latency)

	if len(d.list) > latencyHistorySize {
		d.list = d.list[1:]
	}

	d.sum.total += latency.total
	d.sum.clientServer += latency.clientServer
	d.sum.serverServer += latency.serverServer
	d.sum.serverClient += latency.serverClient

	d.count++

	draw()
}

func draw() {
	width, height := screen.Size()

	blankLine := strings.Repeat(" ", width)

	drawText(0, 0, frameStyle, blankLine)
	drawTextFull(0, 0, false, st{oneUserIDStyle, one.userID}, st{twoUserIDStyle, " <-"}, st{oneUserIDStyle, "-> "}, st{twoUserIDStyle, two.userID})
	drawTextFull(width, 0, true, st{frameStyle, "press "}, st{frameBrightStyle, "Esc"}, st{frameStyle, " to quit"})
	drawText(0, 1, graphStyle, strings.Repeat("\u2594", width))

	for y := 2; y < height-4; y++ {
		drawText(0, y, graphStyle, blankLine)
	}

	drawGraph(0, 2, width, height-6)

	drawText(0, height-4, graphStyle, strings.Repeat("\u2581", width))

	for y := height - 3; y < height; y++ {
		drawText(0, y, frameStyle, blankLine)
	}

	drawText(0, height-3, frameStyle, "         total time          client -> server       server -> server       server -> client")

	line := st{frameStyle, strings.Repeat(fmt.Sprintf("now %v  avg %v   ", placeholder, placeholder), 4)}

	drawTextFull(0, height-2, false, st{oneUserIDStyle, "->  "}, line)
	drawTextFull(0, height-1, false, st{twoUserIDStyle, "<-  "}, line)

	for i, data := range []latencyData{twoData, oneData} {
		if data.count == 0 {
			continue
		}

		lastLatency := data.list[len(data.list)-1]

		meanLatency := latency{
			total:        data.sum.total / time.Duration(data.count),
			clientServer: data.sum.clientServer / time.Duration(data.count),
			serverServer: data.sum.serverServer / time.Duration(data.count),
			serverClient: data.sum.serverClient / time.Duration(data.count),
		}

		y := height - 2 + i

		drawText(8, y, frameBrightStyle, formatLatency(lastLatency.total, false))
		drawText(19, y, frameBrightStyle, formatLatency(meanLatency.total, false))

		if latencyBreakdownValid {
			drawText(31, y, frameBrightStyle, formatLatency(lastLatency.clientServer, false))
			drawText(42, y, frameBrightStyle, formatLatency(meanLatency.clientServer, false))
			drawText(54, y, frameBrightStyle, formatLatency(lastLatency.serverServer, false))
			drawText(65, y, frameBrightStyle, formatLatency(meanLatency.serverServer, false))
			drawText(77, y, frameBrightStyle, formatLatency(lastLatency.serverClient, false))
			drawText(88, y, frameBrightStyle, formatLatency(meanLatency.serverClient, false))
		}
	}

	screen.Show()
}

func drawText(x, y int, style tcell.Style, text string) {
	for _, char := range text {
		screen.SetContent(x, y, char, nil, style)
		x++
	}
}

func drawTextFull(x, y int, alignRight bool, parts ...st) {
	if alignRight {
		for _, part := range parts {
			x -= utf8.RuneCountInString(part.text)
		}
	}
	for _, part := range parts {
		drawText(x, y, part.style, part.text)
		x += utf8.RuneCountInString(part.text)
	}
}

func drawGraph(x, y, width, height int) {
	if width < 8 || height < 1 {
		return
	}

	// Make space for vertical axis and labels
	x += 7
	width -= 7

	drawTextFull(x-3, y+height-1, false, st{graphBrightStyle, "0 "}, st{graphStyle.Underline(true), "\u2503"}, st{graphStyle, strings.Repeat("\u2581", width)})

	for i := 1; i < height; i++ {
		if i%2 == 0 {
			drawText(x-7, y+height-1-i, graphStyle, fmt.Sprintf("%v \u2542%v", placeholder, strings.Repeat("\u254C", width)))
		} else {
			screen.SetContent(x-1, y+height-1-i, '\u2503', nil, graphStyle)
		}
	}

	if twoData.count < 1 {
		return
	}

	var bars []bar
	var maxLatency time.Duration

	// The latest bar belongs to receiver one if both receivers
	// have received the same number of messages. This is because
	// receiver two receives the first message, so receiver one
	// is always "catching up".
	styleOne := (oneData.count == twoData.count)

	for i := 0; i < width; i++ {
		position := i / 2

		var latency latency
		if styleOne && position < len(oneData.list) {
			latency = oneData.list[len(oneData.list)-1-position]
		} else if !styleOne && position < len(twoData.list) {
			latency = twoData.list[len(twoData.list)-1-position]
		}

		bars = append([]bar{{styleOne, latency}}, bars...)

		if latency.total > maxLatency {
			maxLatency = latency.total
		}

		styleOne = !styleOne
	}

	// See https://stackoverflow.com/a/2745086
	cellUnit := (maxLatency + time.Duration(height) - 1) / time.Duration(height)

	for i := 2; i < height; i += 2 {
		drawText(x-7, y+height-1-i, graphBrightStyle, formatLatency(time.Duration(i)*cellUnit+(cellUnit/2), true))
	}

	for i, bar := range bars {
		var style tcell.Style
		if bar.styleOne {
			style = oneBarStyle
		} else {
			style = twoBarStyle
		}

		for j := 0; j < height; j++ {
			index := 0

			if bar.latency.total >= (time.Duration(j)+1)*cellUnit {
				index = len(barCharacters) - 1
			} else if bar.latency.total > time.Duration(j)*cellUnit {
				fraction := float64(bar.latency.total-time.Duration(j)*cellUnit) / float64(cellUnit)
				index = int(math.Round(fraction * float64(len(barCharacters)-1)))
			}

			if index > 0 {
				screen.SetContent(x+i, y+height-1-j, barCharacters[index], nil, style)
			}
		}
	}
}

func formatLatency(latency time.Duration, alignRight bool) string {
	milliseconds := latency.Round(time.Millisecond).Milliseconds()
	var text string
	if milliseconds < 1000 {
		text = fmt.Sprintf("%vms", milliseconds)
	} else {
		text = fmt.Sprintf("%.3gs", float64(milliseconds)/1000)
	}
	if len(text) < 5 {
		padding := strings.Repeat(" ", 5-len(text))
		if alignRight {
			text = padding + text
		} else {
			text = text + padding
		}
	}
	return text[:5]
}
