package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	fsys = flag.String("s", "usr1", "kill syscall sent")
	fper = flag.Float64("p", 90, "percentage threshold before sending a kill signal")
	fdur = flag.Duration("d", time.Duration(time.Minute), "duration")
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
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
	for _, p := range pd {
		for _, np := range npd {
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

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("Arguments needed")
	}
	if sysc[*fsys] == 0 {
		log.Fatal("Signal not recognised")
	}
	log.Printf("Looking for commands containing: %q\n", flag.Args())
	log.Printf("Duration fixed to: %v\n", *fdur)
	log.Printf("Looking for %%CPU usage higher than %v%%\n", *fper)
	log.Printf("Kill signal to send: %q\n\n", sysc[*fsys])

	for {
		pd := checkProcs(flag.Args()...)
		time.Sleep(*fdur)
		npd := checkProcs(flag.Args()...)

		killProcs(pd, npd)
		log.Println("[CLEANING] stored PIDs\n")
		time.Sleep(*fdur)
	}
}
