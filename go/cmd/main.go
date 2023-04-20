package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"fusefs"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
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
	mountpoint := flag.Arg(0)

	log.Println("Mounting FUSE filesystem")
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("fusefs"),
		fuse.Subtype("fusefs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Println("Closing FUSE connection")
		c.Close()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for range sigChan {
			log.Println("Unmounting FUSE filesystem")
			fuse.Unmount(mountpoint)
		}
	}()

	server := fs.New(c, &fs.Config{Debug: func(msg interface{}) { fmt.Println(msg) }})
	err = server.Serve(fusefs.NewFuseFS())
	if err != nil {
		log.Fatal(err)
	}
}
