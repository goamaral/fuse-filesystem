package main

import (
	"flag"
	"fmt"
	fusefs "fusefs/internal"
	"log"
	"os"
	"os/signal"
)

func usage() {
	fmt.Printf("Usage: %s MOUNT_POINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}

	fs := fusefs.NewFuseFS(flag.Arg(0))
	err := fs.Mount()
	if err != nil {
		log.Fatalln("MountErr", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		err := fs.Unmount()
		if err != nil {
			log.Println("UnmountErr", err)
		}
	}()

	err = fs.Serve()
	if err != nil {
		log.Fatalln("ServeErr", err)
	}
}
