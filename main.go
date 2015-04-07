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
	fdur = flag.Duration("i", time.Duration(time.Minute), "interval of each check")

	selfPID int = os.Getpid()
)

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("Arguments needed")
	}
	if sysc[*fsys] == 0 {
		log.Fatal("Signal not recognised")
	}
	log.Printf("Looking for commands containing: %q\n", flag.Args())
	log.Printf("Interval fixed to: %v\n", *fdur)
	log.Printf("Looking for %%CPU usage higher than %v%%\n", *fper)
	log.Printf("Kill signal to send: %q\n\n", sysc[*fsys])

	var pd, npd []uint64
	for {
		pd = checkProcs(flag.Args()...)

		killProcs(pd, npd)
		time.Sleep(*fdur)
		npd = pd
	}
}
