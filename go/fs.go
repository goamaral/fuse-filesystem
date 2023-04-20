package fusefs

import (
	"os"

	"bazil.org/fuse/fs"
)

func NewFuseFS() FuseFS {
	rfs := &fuseFS{LastInode: 1}
	rfs.RootNode = &fuseFSNode{
		FS:    rfs,
		Inode: 1,
		Mode:  os.ModeDir | 0o555,
	}
	return rfs
}

type FuseFS interface {
	fs.FS
	// fs.FSStatfser
	// fs.FSDestroyer
	fs.FSInodeGenerator
}

type fuseFS struct {
	RootNode  FuseFSNode
	LastInode uint64
}

func (rfs fuseFS) GenerateInode(parentInode uint64, name string) uint64 {
	rfs.LastInode++
	return rfs.LastInode
}

func (rfs fuseFS) Root() (fs.Node, error) {
	return rfs.RootNode, nil
}
