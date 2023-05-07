package fusefs

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FuseFSNode interface {
	fs.Node
	// fs.NodeGetattrer
	// fs.NodeSetattrer
	// fs.NodeSymlinker
	// fs.NodeReadlinker
	// fs.NodeLinker
	// fs.NodeRemover
	// fs.NodeAccesser
	fs.NodeStringLookuper
	fs.NodeMkdirer
	fs.NodeOpener
	fs.NodeCreater
	// fs.NodeForgetter
	// fs.NodeRenamer
	// fs.NodeMknoder
	// fs.NodeFsyncer
	// fs.NodeGetxattrer
	// fs.NodeListxattrer
	// fs.NodeSetxattrer
	// fs.NodeRemovexattrer

	fs.HandleWriter
}

func NewFuseFSNode() FuseFSNode {
	return &fuseFSNode{}
}

type fuseFSNode struct {
	FS     FuseFS
	Name   string
	Inode  uint64
	Mode   os.FileMode
	Nodes  []*fuseFSNode
	IsOpen bool
	Data   []byte
}

func (n fuseFSNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = n.Inode
	attr.Mode = n.Mode
	return nil
}

func (n fuseFSNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
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

func (n *fuseFSNode) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if !n.Mode.IsDir() {
		return nil, nil, syscall.ENOTDIR
	}
	newNode := &fuseFSNode{
		FS:    n.FS,
		Name:  req.Name,
		Inode: n.FS.GenerateInode(n.Inode, req.Name),
		Mode:  req.Mode,
	}
	n.Nodes = append(n.Nodes, newNode)
	return newNode, nil, nil
}

func (n *fuseFSNode) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
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

func (n *fuseFSNode) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	if !n.Mode.IsDir() {
		return nil, syscall.ENOTDIR
	}
	newNode := &fuseFSNode{
		FS:    n.FS,
		Name:  req.Name,
		Inode: n.FS.GenerateInode(n.Inode, req.Name),
		Mode:  req.Mode,
	}
	n.Nodes = append(n.Nodes, newNode)
	return newNode, nil
}

func (n *fuseFSNode) Open(ctx context.Context, req *fuse.OpenRequest, res *fuse.OpenResponse) (fs.Handle, error) {
	n.IsOpen = true
	return n, nil
}

/*
Ignored fields

	type WriteRequest struct {
		Handle    HandleID
		Offset    int64
		Flags     WriteFlags
		LockOwner LockOwner
		FileFlags OpenFlags
	}

	type WriteResponse struct {
		Size int
	}
*/
func (n *fuseFSNode) Write(ctx context.Context, req *fuse.WriteRequest, res *fuse.WriteResponse) error {
	if !n.IsOpen {
		return syscall.EBADF
	}
	if req.FileFlags.IsReadOnly() {
		return syscall.EBADF
	}

	n.Data = req.Data
	res.Size = len(req.Data)

	return nil
}
