package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

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

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("fusefs"),
		fuse.Subtype("fusefs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	server := fs.New(c, &fs.Config{Debug: func(msg interface{}) { fmt.Println(msg) }})
	err = server.Serve(NewFuseFS())
	if err != nil {
		log.Fatal(err)
	}
}

func NewFuseFS() *FuseFS {
	rfs := &FuseFS{LastInode: 1}
	rfs.RootNode = &FuseFSNode{
		FS:    rfs,
		Inode: 1,
		Mode:  os.ModeDir | 0o555,
	}
	return rfs
}

type FuseFS struct {
	RootNode  *FuseFSNode
	LastInode uint64
}

func (rfs FuseFS) GetNextInode() uint64 {
	rfs.LastInode++
	return rfs.LastInode
}

func (rfs FuseFS) Root() (fs.Node, error) {
	return rfs.RootNode, nil
}

type FuseFSNode struct {
	FS    *FuseFS
	Name  string
	Inode uint64
	Mode  os.FileMode
	Nodes []*FuseFSNode
}

func (n FuseFSNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = n.Inode
	attr.Mode = n.Mode
	return nil
}

func (n FuseFSNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, n := range n.Nodes {
		if n.Name == name {
			return n, nil
		} else if n.Mode.IsDir() {
			if lookupNode, err := n.Lookup(ctx, name); err == nil {
				return lookupNode, nil
			}
		}
	}
	return nil, syscall.ENOENT
}

func (n *FuseFSNode) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if !n.Mode.IsDir() {
		return nil, nil, syscall.ENOTDIR
	}
	newNode := &FuseFSNode{
		FS:    n.FS,
		Name:  req.Name,
		Inode: n.FS.GetNextInode(),
		Mode:  req.Mode,
	}
	n.Nodes = append(n.Nodes, newNode)
	return newNode, nil, nil
}

func (n *FuseFSNode) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	ents := make([]fuse.Dirent, len(n.Nodes))
	for i, node := range n.Nodes {
		typ := fuse.DT_File
		if node.Mode.IsDir() {
			typ = fuse.DT_Dir
		}
		ents[i] = fuse.Dirent{Inode: node.Inode, Type: typ, Name: node.Name}
	}
	return ents, nil
}

func (n *FuseFSNode) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	if !n.Mode.IsDir() {
		return nil, syscall.ENOTDIR
	}
	newNode := &FuseFSNode{
		FS:    n.FS,
		Name:  req.Name,
		Inode: n.FS.GetNextInode(),
		Mode:  req.Mode,
	}
	n.Nodes = append(n.Nodes, newNode)
	return newNode, nil
}
