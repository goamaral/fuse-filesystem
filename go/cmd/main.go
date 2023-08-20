package main

import (
	"flag"
	"fmt"
	fusefs "fusefs/internal"
	"os"
	"os/signal"
	"syscall"
)

func usage() {
	fmt.Printf("Usage: %s MOUNT_POINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	nArgs := flag.NArg()
	if nArgs >= 1 {
		usage()
		os.Exit(2)
	}

	path := "/tmp/fusefs"
	if nArgs == 1 {
		path = flag.Arg(0)
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0700)
		if err != nil {
			fusefs.Fatalf("failed to create %s: %v", path, err)
		}
		defer func() {
			if err = os.Remove(path); err != nil {
				fusefs.Fatalf("failed to remove %s: %v", path, err)
			}
		}()
	} else if err != nil {
		fusefs.Fatalf("failed to check if %s exists: %v", path, err)
	}
	fusefs.Infof("Mounted filesystem at %s", path)

	fs := fusefs.NewFuseFS(path)
	if err = fs.Mount(); err != nil {
		fusefs.Fatalf("failed to mount: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-sigChan
		fusefs.Infof("Unmounted filesystem from %s", path)
		if err = fs.Unmount(); err != nil {
			fusefs.Fatalf("failed to unmount: %w", err)
		}
	}()

	if err = fs.Serve(); err != nil {
		fusefs.Fatalf("failed to serve: %w", err)
	}
}
