package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"

	"fmt"
	"sort"
	"time"
)

func (hi *Hi) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*Vnet)
	if !in.Parse("%v", v.hwIfIndexByName, hi) {
		panic(parse.ErrInput)
	}
}

func (si *Si) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*Vnet)
	var hi Hi
	if !in.Parse("%v", v.hwIfIndexByName, &hi) {
		panic(parse.ErrInput)
	}
	// Initially get software interface from hardware interface.
	hw := v.HwIf(hi)
	*si = hw.si
	var (
		id IfIndex
		ok bool
	)
	if in.Parse(".%d", &id) {
		if *si, ok = hw.subSiById[id]; !ok {
			panic(fmt.Errorf("unkown sub interface id: %d", id))
		}
	}
}

type showIfConfig struct {
	detail bool
	re     parse.Regexp
	colMap map[string]bool
	siMap  map[Si]bool
	hiMap  map[Hi]bool
}

func (c *showIfConfig) parse(v *Vnet, in *cli.Input, isHw bool) {
	c.detail = false
	c.colMap = map[string]bool{
		"Rate": false,
	}
	if isHw {
		c.hiMap = make(map[Hi]bool)
	} else {
		c.siMap = make(map[Si]bool)
	}
	for !in.End() {
		var (
			si Si
			hi Hi
		)
		switch {
		case !isHw && in.Parse("%v", &si, v):
			c.siMap[si] = true
		case isHw && in.Parse("%v", &hi, v):
			c.hiMap[hi] = true
		case in.Parse("m%*atching %v", &c.re):
		case in.Parse("d%*etail"):
			c.detail = true
		case in.Parse("r%*ate"):
			c.colMap["Rate"] = true
		default:
			panic(parse.ErrInput)
		}
	}
}

type swIfIndices struct {
	*Vnet
	ifs []Si
}

func (h *swIfIndices) Less(i, j int) bool { return h.SwLessThan(h.SwIf(h.ifs[i]), h.SwIf(h.ifs[j])) }
func (h *swIfIndices) Swap(i, j int)      { h.ifs[i], h.ifs[j] = h.ifs[j], h.ifs[i] }
func (h *swIfIndices) Len() int           { return len(h.ifs) }

type showSwIf struct {
	Name    string `format:"%-30s" align:"left"`
	State   string `format:"%-12s" align:"left"`
	Counter string `format:"%-30s" align:"left"`
	Count   string `format:"%16s" align:"right"`
	Rate    string `format:"%16s" align:"right"`
}
type showSwIfs []showSwIf

func (v *Vnet) showSwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {

	cf := &showIfConfig{}
	cf.parse(v, in, false)

	swIfs := &swIfIndices{Vnet: v}
	if len(cf.siMap) == 0 {
		for i := range v.swInterfaces.elts {
			si := Si(i)
			if v.swInterfaces.IsFree(uint(si)) {
				continue
			}
			if cf.re.Valid() && !cf.re.MatchString(si.Name(v)) {
				continue
			}
			swIfs.ifs = append(swIfs.ifs, si)
		}
	} else {
		for si, _ := range cf.siMap {
			swIfs.ifs = append(swIfs.ifs, si)
		}
	}

	if cf.re.Valid() && len(swIfs.ifs) == 0 {
		fmt.Fprintf(w, "No interfaces match expression: `%s'\n", cf.re)
		return
	}

	sort.Sort(swIfs)

	v.syncSwIfCounters()

	sifs := showSwIfs{}
	dt := time.Since(v.timeLastClear).Seconds()
	alwaysReport := len(cf.siMap) > 0 || cf.re.Valid()
	for i := range swIfs.ifs {
		si := v.SwIf(swIfs.ifs[i])
		first := true
		firstIf := showSwIf{
			Name:  si.IfName(v),
			State: si.flags.String(),
		}
		v.foreachSwIfCounter(cf.detail, si.si, func(counter string, count uint64) {
			s := showSwIf{
				Counter: counter,
				Count:   fmt.Sprintf("%d", count),
				Rate:    fmt.Sprintf("%.2e", float64(count)/dt),
			}
			if first {
				first = false
				s.Name = firstIf.Name
				s.State = firstIf.State
			}
			sifs = append(sifs, s)
		})
		// Always at least report name and state for specified interfaces.
		if first && alwaysReport {
			sifs = append(sifs, firstIf)
		}
	}
	if len(sifs) > 0 {
		elib.Tabulate(sifs).WriteCols(w, cf.colMap)
	} else {
		fmt.Fprintln(w, "All counters are zero")
	}
	return
}

func (v *Vnet) clearSwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	v.clearIfCounters()
	return
}

type hwIfIndices struct {
	*Vnet
	ifs []Hi
}

func (h *hwIfIndices) Less(i, j int) bool { return h.HwLessThan(h.HwIf(h.ifs[i]), h.HwIf(h.ifs[j])) }
func (h *hwIfIndices) Swap(i, j int)      { h.ifs[i], h.ifs[j] = h.ifs[j], h.ifs[i] }
func (h *hwIfIndices) Len() int           { return len(h.ifs) }

type showHwIf struct {
	Name    string `format:"%-30s"`
	Address string `format:"%-12s" align:"center"`
	Link    string `width:12`
	Counter string `format:"%-30s" align:"left"`
	Count   string `format:"%16s" align:"right"`
	Rate    string `format:"%16s" align:"right"`
}
type showHwIfs []showHwIf

func (ns showHwIfs) Less(i, j int) bool { return ns[i].Name < ns[j].Name }
func (ns showHwIfs) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns showHwIfs) Len() int           { return len(ns) }

func (v *Vnet) showHwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	cf := showIfConfig{}
	cf.parse(v, in, true)

	hwIfs := &hwIfIndices{Vnet: v}

	if len(cf.hiMap) == 0 {
		for i := range v.hwIferPool.elts {
			if v.hwIferPool.IsFree(uint(i)) {
				continue
			}
			h := v.hwIferPool.elts[i].GetHwIf()
			if h.unprovisioned {
				continue
			}
			if cf.re.Valid() && !cf.re.MatchString(h.name) {
				continue
			}
			hwIfs.ifs = append(hwIfs.ifs, Hi(i))
		}
	} else {
		for hi, _ := range cf.hiMap {
			hwIfs.ifs = append(hwIfs.ifs, hi)
		}
	}

	if cf.re.Valid() && len(hwIfs.ifs) == 0 {
		fmt.Fprintf(w, "No interfaces match expression: `%s'\n", cf.re)
		return
	}

	sort.Sort(hwIfs)

	ifs := showHwIfs{}
	dt := time.Since(v.timeLastClear).Seconds()
	alwaysReport := len(cf.siMap) > 0 || cf.re.Valid()
	for i := range hwIfs.ifs {
		hi := v.HwIfer(hwIfs.ifs[i])
		h := hi.GetHwIf()
		first := true
		firstIf := showHwIf{
			Name:    h.name,
			Address: hi.FormatAddress(),
			Link:    h.LinkString(),
		}
		v.foreachHwIfCounter(cf.detail, h.hi, func(counter string, count uint64) {
			s := showHwIf{
				Counter: counter,
				Count:   fmt.Sprintf("%d", count),
				Rate:    fmt.Sprintf("%.2e", float64(count)/dt),
			}
			if first {
				first = false
				s.Name = firstIf.Name
				s.Address = firstIf.Address
				s.Link = firstIf.Link
			}
			ifs = append(ifs, s)
		})
		// Always at least report name and state for specified interfaces.
		if first && alwaysReport {
			ifs = append(ifs, firstIf)
		}
	}
	if len(ifs) > 0 {
		elib.Tabulate(ifs).WriteCols(w, cf.colMap)
	} else {
		fmt.Println("All counters are zero")
	}
	return
}

func (v *Vnet) setSwIf(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var (
		isUp parse.UpDown
		si   Si
	)
	switch {
	case in.Parse("state %v %v", &si, v, &isUp):
		s := v.SwIf(si)
		err = s.SetAdminUp(v, bool(isUp))
	default:
		err = cli.ParseError
	}
	return
}

func (v *Vnet) setHwIf(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var hi Hi

	var (
		mtu       uint
		bw        Bandwidth
		provision parse.Enable
		loopback  IfLoopbackType
	)

	switch {
	case in.Parse("l%*oopback %v %v", &hi, v, &loopback):
		h := v.HwIfer(hi)
		err = h.SetLoopback(loopback)
	case in.Parse("mtu %v %d", &hi, v, &mtu):
		h := v.HwIf(hi)
		err = h.SetMaxPacketSize(mtu)
	case in.Parse("p%*rovision %v %v", &hi, v, &provision):
		h := v.HwIf(hi)
		err = h.SetProvisioned(bool(provision))
	case in.Parse("s%*peed %v %v", &hi, v, &bw):
		h := v.HwIf(hi)
		err = h.SetSpeed(bw)
	default:
		err = cli.ParseError
	}
	return
}

func init() {
	AddInit(func(v *Vnet) {
		cmds := [...]cli.Command{
			cli.Command{
				Name:      "show interfaces",
				ShortHelp: "show interface statistics",
				Action:    v.showSwIfs,
			},
			cli.Command{
				Name:      "clear interfaces",
				ShortHelp: "clear interface statistics",
				Action:    v.clearSwIfs,
			},
			cli.Command{
				Name:      "show hardware-interfaces",
				ShortHelp: "show hardware interface statistics",
				Action:    v.showHwIfs,
			},
			cli.Command{
				Name:      "set interface",
				ShortHelp: "set interface commands",
				Action:    v.setSwIf,
			},
			cli.Command{
				Name:      "set hardware-interface",
				ShortHelp: "set hardware interface commands",
				Action:    v.setHwIf,
			},
			cli.Command{
				Name:      "show buffers",
				ShortHelp: "show dma buffer usage",
				Action:    v.showBufferUsage,
			},
		}
		for i := range cmds {
			v.CliAdd(&cmds[i])
		}
	})
}