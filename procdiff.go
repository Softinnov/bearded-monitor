package main

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/c9s/goprocinfo/linux"
)

type ProcDiff []*Proc

func (pd *ProcDiff) Contains(cmd ...string) {

	maxp, e := linux.ReadMaxPID("/proc/sys/kernel/pid_max")
	if e != nil {
		log.Println(e)
		return
	}
	for p := uint64(1); p < maxp; p++ {
		// Do not check it's own PID
		if int(p) == selfPID {
			continue
		}
		sp := strconv.FormatUint(p, 10)
		status, e := linux.ReadProcessStatus(filepath.Join("/proc", sp, "status"))
		if e != nil {
			continue
		}
		cmdline, e := linux.ReadProcessCmdline(filepath.Join("/proc", sp, "cmdline"))
		if e != nil {
			continue
		}
		// Tgid != Pid for a thread
		if status.Tgid != status.Pid {
			continue
		}
		for _, c := range cmd {
			if strings.Contains(cmdline, c) {
				*pd = append(*pd, &Proc{p, cmdline, nil, nil, 0})
				break
			}
		}
	}
}

func cpu() uint64 {

	stat, e := linux.ReadStat("/proc/stat")
	if e != nil {
		log.Println(e)
		return 0
	}
	s := stat.CPUStatAll
	return s.User + s.Nice + s.System + s.Idle
}

func (pd *ProcDiff) Percentage() {

	for _, p := range *pd {
		sp := strconv.FormatUint(p.Pid, 10)
		r, e := linux.ReadProcessStat(filepath.Join("/proc", sp, "stat"))
		if e != nil {
			log.Println(e)
			continue
		}
		p.Fir = r
	}
	cpu1 := cpu()

	time.Sleep(time.Second)

	for _, p := range *pd {
		sp := strconv.FormatUint(p.Pid, 10)
		r, e := linux.ReadProcessStat(filepath.Join("/proc", sp, "stat"))
		if e != nil {
			log.Println(e)
			continue
		}
		p.Sec = r
	}

	cpu2 := cpu()

	for _, p := range *pd {

		p1 := p.Fir
		p2 := p.Sec
		if p1 == nil || p2 == nil {
			continue
		}
		user := int64(p2.Utime-p1.Utime) + (p2.Cutime - p1.Cutime)
		system := int64(p2.Stime-p1.Stime) + (p2.Cstime - p1.Cstime)

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
			log.Printf("[WARNING]: %v: %v%% (%v)", p.Pid, math.Trunc(p.Per), p.Cmd)
			pdf = append(pdf, p.Pid)
		} else if *fver == true {
			log.Printf("%v: %v%% (%v)", p.Pid, math.Trunc(p.Per), p.Cmd)
		}
	}
	if *fver == true {
		log.Printf("Found \033[1m%v\033[0m corresponding processes, with \033[1m%v\033[0m > %v%%.\n", len(*pd), len(pdf), *fper)
	}

	return pdf
}

func killProcs(pd, npd []uint64) {
	for _, np := range npd {
		for _, p := range pd {
			if p != np {
				continue
			}
			proc, e := os.FindProcess(int(p))
			if e != nil {
				log.Println(e)
			}
			log.Printf("[KILL] sent signal %s to %v\n", *fsys, proc.Pid)
			e = proc.Signal(sysc[*fsys])
			if e != nil {
				log.Println(e)
			}
		}
	}
}
