package main

import "github.com/c9s/goprocinfo/linux"

type Proc struct {
	Pid uint64
	Cmd string
	Fir *linux.ProcessStat
	Sec *linux.ProcessStat
	Per float64
}
