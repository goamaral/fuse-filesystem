package fusefs

import (
	"fmt"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func NewFuseFS(mountpoint string) FuseFS {
	rfs := &fuseFS{Mountpoint: mountpoint, LastInode: 1}
	rfs.RootNode = &fuseFSNode{
		FS:    rfs,
		Inode: 1,
		Mode:  os.ModeDir | 0o555,
	}
	return rfs
}

type FuseFS interface {
	Mount() error
	Serve() error
	Unmount() error

	fs.FS
	// fs.FSStatfser
	// fs.FSDestroyer
	fs.FSInodeGenerator
}

type fuseFS struct {
	Mountpoint string
	Conn       *fuse.Conn
	RootNode   FuseFSNode
	LastInode  uint64
}

func (rfs *fuseFS) Mount() error {
	c, err := fuse.Mount(
		rfs.Mountpoint,
		fuse.FSName("fusefs"),
		fuse.Subtype("fusefs"),
	)
	if err != nil {
		return err
	}
	rfs.Conn = c

	return nil
}

func (rfs *fuseFS) Serve() error {
	server := fs.New(rfs.Conn, &fs.Config{Debug: func(msg interface{}) { fmt.Println(msg) }})
	return server.Serve(rfs)
}

func (rfs *fuseFS) Unmount() error {
	err := fuse.Unmount(rfs.Mountpoint)
	if err != nil {
		return err
	}

	return rfs.Conn.Close()
}

func (rfs fuseFS) GenerateInode(parentInode uint64, name string) uint64 {
	rfs.LastInode++
	return rfs.LastInode
}

func (rfs fuseFS) Root() (fs.Node, error) {
	return rfs.RootNode, nil
}
