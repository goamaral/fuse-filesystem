package main

import (
	"flag"
	"fmt"
	fusefs "fusefs/internal"
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

	nArgs := flag.NArg()
	if nArgs >= 1 {
		usage()
		os.Exit(2)
	}

	var path string
	var err error
	if nArgs == 0 {
		path, err = os.MkdirTemp("/tmp", "fusefs-")
		if err != nil {
			fusefs.Fatalf("failed to create temp dir: %w", err)
		}
		fusefs.Infof("Created temporary filesystem at %s", path)

		defer func() {
			err := os.Remove(path)
			if err != nil {
				fusefs.Fatalf("failed to remove temp dir: %w", err)
			}
			fusefs.Infof("Removed temporary filesystem from %s", path)
		}()
	} else if nArgs == 1 {
		path = flag.Arg(0)
	}

	fs := fusefs.NewFuseFS(path)
	err = fs.Mount()
	if err != nil {
		fusefs.Fatalf("failed to mount: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		err := fs.Unmount()
		if err != nil {
			fusefs.Fatalf("failed to unmount: %w", err)
		}
	}()

	err = fs.Serve()
	if err != nil {
		fusefs.Fatalf("failed to serve: %w", err)
	}
}
