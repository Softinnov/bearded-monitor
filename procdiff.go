package main

import (
	"log"
	"os"
	"runtime"
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
			// Tgid != Pid for a thread
			if strings.Contains(proc.Cmdline, c) && proc.Status.Tgid == proc.Status.Pid {
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

		p.Per = (float64(user+system) / float64((cpu2-cpu1)/uint64(runtime.NumCPU()))) * 100
	}
}

func checkProcs(ss ...string) []uint64 {
	pd := new(ProcDiff)

	pd.Contains(ss...)
	pd.Percentage()

	pdf := []uint64{}

	for _, p := range *pd {
		if p.Per > *fper {
			log.Printf("[WARNING]: %v: %v%% (%v)", p.Pid, p.Per, p.Cmd)
			pdf = append(pdf, p.Pid)
		}
	}
	log.Printf("Found \033[1m%v\033[0m corresponding processes, with \033[1m%v\033[0m > %v%%.\n", len(*pd), len(pdf), *fper)

	return pdf
}

func killProcs(pd, npd []uint64) {
	for _, np := range npd {
		for _, p := range pd {
			if p != np {
				continue
			}
			proc, e := os.FindProcess(int(p))
			check(e)
			log.Printf("[KILL] sent signal %s to %v\n", *fsys, proc.Pid)
			e = proc.Signal(sysc[*fsys])
			check(e)
		}
	}
}
