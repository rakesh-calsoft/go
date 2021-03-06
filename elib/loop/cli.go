// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/elib/parse"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type LoopCli struct {
	Node
	cli.Main
}

func (l *Loop) CliAdd(c *cli.Command) { l.Cli.AddCommand(c) }

type fileEvent struct {
	c *LoopCli
	*cli.File
}

func (c *LoopCli) rxReady(f *cli.File) {
	c.AddEvent(&fileEvent{c: c, File: f}, c)
}

func (c *fileEvent) EventAction() {
	if err := c.RxReady(); err == cli.ErrQuit {
		c.c.AddEvent(ErrQuit, nil)
	}
}

func (c *fileEvent) String() string { return "rx-ready " + c.File.String() }

func (c *LoopCli) EventHandler() {}

func (c *LoopCli) LoopInit(l *Loop) {
	if len(c.Main.Prompt) == 0 {
		c.Main.Prompt = "cli# "
	}
	c.Main.Start()
}

func (c *LoopCli) LoopExit(l *Loop) {
	c.Main.End()
}

type loggerMain struct {
	once sync.Once
	w    io.Writer
	l    *log.Logger
}

func (l *Loop) loggerInit() {
	m := &l.loggerMain
	if m.w = l.LogWriter; m.w == nil {
		m.w = os.Stdout
	}
	m.l = log.New(m.w, "", log.Lmicroseconds)
	return
}

func (l *Loop) Logf(format string, args ...interface{}) {
	m := &l.loggerMain
	m.once.Do(l.loggerInit)
	m.l.Printf(format, args...)
}
func (l *Loop) Logln(args ...interface{}) {
	m := &l.loggerMain
	m.once.Do(l.loggerInit)
	m.l.Println(args...)
}
func (m *loggerMain) Fatalf(format string, args ...interface{}) { panic(fmt.Errorf(format, args...)) }

type rtNode struct {
	Name     string  `format:"%-30s"`
	State    string  `align:"center"`
	Calls    uint64  `format:"%16d"`
	Vectors  uint64  `format:"%16d"`
	Suspends uint64  `format:"%16d"`
	Clocks   float64 `format:"%16.2f"`
}
type rtNodes []rtNode

func (ns rtNodes) Less(i, j int) bool { return ns[i].Name < ns[j].Name }
func (ns rtNodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns rtNodes) Len() int           { return len(ns) }

func (l *Loop) showRuntimeStats(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	colMap := map[string]bool{
		"State": false,
	}
	show_detail := false
	for !in.End() {
		switch {
		case in.Parse("d%*etail"):
			colMap["State"] = true
			show_detail = true
		default:
			panic(parse.ErrInput)
		}
	}

	ns := rtNodes{}
	for i := range l.DataNodes {
		n := l.DataNodes[i].GetNode()
		var s [2]stats
		s[0].add(&n.inputStats)
		s[1].add(&n.outputStats)
		l.activePollerPool.Foreach(func(a *activePoller) {
			if a.activeNodes != nil {
				s[0].add(&a.activeNodes[i].inputStats)
				s[1].add(&a.activeNodes[i].outputStats)
			}
		})
		name := n.name
		_, isIn := n.noder.(inLooper)
		_, isOut := n.noder.(outLooper)
		_, isInOut := n.noder.(inOutLooper)
		for j := range s {
			if j == 0 && !isIn && !isInOut {
				continue
			}
			if j == 1 && !isOut && !isInOut {
				continue
			}
			io := ""
			if (isIn && isOut) || isInOut {
				if j == 0 {
					io = " in"
				} else {
					io = " out"
				}
			}
			if s[j].calls > 0 || show_detail {
				state := ""
				if j == 0 {
					state = fmt.Sprintf("%s", n.get_flags())
				}
				ns = append(ns, rtNode{
					Name:     name + io,
					State:    state,
					Calls:    s[j].calls,
					Vectors:  s[j].vectors,
					Suspends: s[j].suspends,
					Clocks:   s[j].clocksPerVector(),
				})
			}
		}
	}

	// Summary
	{
		var s stats
		l.activePollerPool.Foreach(func(a *activePoller) {
			s.add(&a.pollerStats)
		})
		if s.calls > 0 {
			dt := time.Since(l.timeLastRuntimeClear).Seconds()
			vecsPerSec := float64(s.vectors) / dt
			clocksPerVec := float64(s.clocks) / float64(s.vectors)
			vecsPerCall := float64(s.vectors) / float64(s.calls)
			fmt.Fprintf(w, "Vectors: %d, Vectors/sec: %.2e, Clocks/vector: %.2f, Vectors/call %.2f\n",
				s.vectors, vecsPerSec, clocksPerVec, vecsPerCall)
		}
	}

	sort.Sort(ns)
	elib.Tabulate(ns).WriteCols(w, colMap)
	return
}

func (l *Loop) clearRuntimeStats(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	l.timeLastRuntimeClear = time.Now()
	for i := range l.DataNodes {
		n := l.DataNodes[i].GetNode()
		n.inputStats.clear()
		n.outputStats.clear()
	}
	l.activePollerPool.Foreach(func(a *activePoller) {
		a.pollerStats.clear()
		for j := range a.activeNodes {
			a.activeNodes[j].inputStats.clear()
			a.activeNodes[j].outputStats.clear()
		}
	})
	return
}

func (l *Loop) showEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	detail := false
	for !in.End() {
		switch {
		case in.Parse("de%*tail"):
			detail = true
		default:
			err = parse.ErrInput
			return
		}
	}
	v := elog.NewView()
	v.Print(w, detail)
	return
}

func (l *Loop) clearEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	elog.Clear()
	return
}

func addDelEventFilter(f string, isDel bool) {
	v := true
	if f[0] == '!' {
		v = false
		f = f[1:]
	}
	elog.AddDelEventFilter(f, v, isDel)
}

func (l *Loop) configEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var filter string
	for !in.End() {
		switch {
		case in.Parse("add %v", &filter):
			addDelEventFilter(filter, false)
		case in.Parse("del %v", &filter):
			addDelEventFilter(filter, true)
		case in.Parse("reset"):
			elog.ResetFilters()
		default:
			err = parse.ErrInput
			return
		}
	}
	return
}

func (l *Loop) exec(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var files []*os.File
	for !in.End() {
		var (
			pattern string
			names   []string
			f       *os.File
		)
		in.Parse("%s", &pattern)
		if names, err = filepath.Glob(pattern); err != nil {
			return
		}
		if len(names) == 0 {
			err = fmt.Errorf("no files matching pattern: `%s'", pattern)
			return
		}
		for _, name := range names {
			if f, err = os.OpenFile(name, os.O_RDONLY, 0); err != nil {
				return
			}
			files = append(files, f)
		}
	}
	for _, f := range files {
		var i [2]cli.Input
		i[0].Init(f)
		for !i[0].End() {
			i[1].Init(nil)
			if !i[0].Parse("%l", &i[1].Input) {
				err = i[0].Error()
				return
			}
			if err = l.Cli.ExecInput(w, &i[1]); err != nil {
				return
			}
		}
		f.Close()
	}
	return
}

func (l *Loop) comment(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	in.Skip()
	return
}

func (l *Loop) cliInit() {
	l.RegisterEventPoller(iomux.Default)
	c := &l.Cli
	c.Main.RxReady = c.rxReady
	l.RegisterNode(c, "loop-cli")
	c.AddCommand(&cli.Command{
		Name:      "show runtime",
		ShortHelp: "show main loop runtime statistics",
		Action:    l.showRuntimeStats,
	})
	c.AddCommand(&cli.Command{
		Name:      "clear runtime",
		ShortHelp: "clear main loop runtime statistics",
		Action:    l.clearRuntimeStats,
	})
	c.AddCommand(&cli.Command{
		Name:      "show event-log",
		ShortHelp: "show events in event log",
		Action:    l.showEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "clear event-log",
		ShortHelp: "clear events in event log",
		Action:    l.clearEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "event-log",
		ShortHelp: "event log commands",
		Action:    l.configEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "exec",
		ShortHelp: "execute cli commands from given file(s)",
		Action:    l.exec,
	})
	c.AddCommand(&cli.Command{
		Name:      "//",
		ShortHelp: "comment",
		Action:    l.comment,
	})
}
