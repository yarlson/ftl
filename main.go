package main

import (
	"fmt"
	"github.com/yarlson/ftl/pkg/console"
	"os"
	"os/signal"
	"syscall"

	"github.com/yarlson/ftl/cmd"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		console.Reset()
		os.Exit(1)
	}()

	defer console.Reset()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
