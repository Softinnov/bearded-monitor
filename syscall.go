package main

import "syscall"

var sysc map[string]syscall.Signal = map[string]syscall.Signal{
	"usr1": syscall.SIGUSR1,
	"usr2": syscall.SIGUSR2,
	"kill": syscall.SIGKILL,
	"term": syscall.SIGTERM,
	"stop": syscall.SIGSTOP,
	"int":  syscall.SIGINT,
	"hup":  syscall.SIGHUP,
}
