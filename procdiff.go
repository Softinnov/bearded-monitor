package main

import (
	"strings"
	"time"

	"github.com/c9s/goprocinfo/linux"
)

type ProcDiff []*Proc

func (pd *ProcDiff) Contains(cmd ...string) {

	maxp, e := linux.ReadMaxPID("/proc/sys/kernel/pid_max")
	check(e)
	for p := uint64(1); p < maxp; p++ {
		// Do not check it's own PID
		if int(p) == selfPID {
			continue
		}
		proc, e := linux.ReadProcess(p, "/proc")
		if e != nil {
			continue
		}
		for _, c := range cmd {
			// ExitSignal to check threads
			if strings.Contains(proc.Cmdline, c) && proc.Stat.ExitSignal != -1 {
				*pd = append(*pd, &Proc{p, proc.Cmdline, nil, nil, 0})
				break
			}
		}
	}
}

func cpu() uint64 {

	cpu, e := linux.ReadStat("/proc/stat")
	check(e)

	c := cpu.CPUStatAll
	return c.User + c.Nice + c.System + c.Idle
}

func (pd *ProcDiff) Percentage() {

	for _, p := range *pd {
		r, e := linux.ReadProcess(p.Pid, "/proc")
		check(e)
		p.Fir = r
	}
	cpu1 := cpu()

	time.Sleep(time.Second)

	for _, p := range *pd {
		r, e := linux.ReadProcess(p.Pid, "/proc")
		check(e)
		p.Sec = r
	}

	cpu2 := cpu()

	for _, p := range *pd {

		p1 := p.Fir
		p2 := p.Sec
		user := int64(p2.Stat.Utime-p1.Stat.Utime) + (p2.Stat.Cutime - p1.Stat.Cutime)
		system := int64(p2.Stat.Stime-p1.Stat.Stime) + (p2.Stat.Cstime - p1.Stat.Cstime)

		p.Per = (float64(user+system) / float64((cpu2-cpu1)/8)) * 100
	}
}
