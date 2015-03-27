package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/c9s/goprocinfo/linux"
)

type ProcDiff []*Proc

type Proc struct {
	Pid uint64
	Cmd string
	Fir *linux.Process
	Sec *linux.Process
	Per float64
}

func (pd *ProcDiff) Contains(cmd ...string) {

	maxp, e := linux.ReadMaxPID("/proc/sys/kernel/pid_max")
	check(e)
	for p := uint64(1); p < maxp; p++ {
		proc, e := linux.ReadProcess(p, "/proc")
		if e != nil {
			continue
		}
		for _, c := range cmd {
			if strings.Contains(proc.Cmdline, c) {
				*pd = append(*pd, &Proc{p, proc.Cmdline, nil, nil, 0})
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

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkProcs() []uint64 {
	pd := new(ProcDiff)

	pd.Contains("loop")
	pd.Percentage()

	pdf := []uint64{}

	for _, p := range *pd {
		if p.Per > 90 {
			log.Printf("WARNING: %v: %v%% (%v)", p.Pid, p.Per, p.Cmd)
			pdf = append(pdf, p.Pid)
		}
	}

	if len(pdf) > 0 {
		b, e := json.Marshal(pdf)
		check(e)
		e = ioutil.WriteFile("pids", b, 0644)
		check(e)
	}
	fmt.Printf("Found \033[1m%v\033[0m corresponding process, with \033[1m%v\033[0m > 90%%.\n", len(*pd), len(pdf))

	return pdf
}

func readProcs() []uint64 {
	pdf := []uint64{}

	b, e := ioutil.ReadFile("pids")
	if e != nil {
		log.Println(e)
		return []uint64{}
	}
	e = json.Unmarshal(b, &pdf)
	check(e)

	return pdf
}

func killProcs(pd, npd []uint64) {
	for _, p := range pd {
		for _, np := range npd {
			if p != np {
				continue
			}
			proc, e := os.FindProcess(int(p))
			check(e)
			fmt.Printf("send SIGUSR1 to %v\n", proc.Pid)
			proc.Signal(syscall.SIGUSR1)
		}
	}
}

func main() {

	pd := readProcs()
	npd := checkProcs()

	killProcs(pd, npd)

}
