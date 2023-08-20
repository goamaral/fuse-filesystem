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
	fs.NodeSetattrer
	// fs.NodeSymlinker
	// fs.NodeReadlinker
	// fs.NodeLinker
	fs.NodeRemover
	// fs.NodeAccesser
	fs.NodeStringLookuper
	fs.NodeMkdirer
	// fs.NodeOpener <-
	fs.NodeCreater
	// fs.NodeForgetter
	// fs.NodeRenamer
	// fs.NodeMknoder
	// fs.NodeFsyncer
	fs.NodeGetxattrer
	// fs.NodeListxattrer
	// fs.NodeSetxattrer
	// fs.NodeRemovexattrer
	// fs.NodePoller // fs.HandlePoller <-

	// fs.HandleFlusher <-
	// fs.HandleReadAller
	fs.HandleReadDirAller
	// fs.HandleReader
	fs.HandleWriter
	// fs.HandleReleaser <-
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
	Data   []byte
	Xattrs map[string]string
}

// fs.Node */
func (n fuseFSNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = n.Inode
	attr.Mode = n.Mode
	return nil
}

// fs.NodeSetattrer
/*
Unused fields
	type SetattrRequest struct {
		Valid  SetattrValid
		Handle HandleID
		Size   uint64
		Atime  time.Time
		Mtime  time.Time
		Ctime  time.Time
		Uid  uint32
		Gid  uint32
	}
*/
func (n *fuseFSNode) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// NOTE: res.Atrr is filled by Attr method

	if req.Mode&os.ModeIrregular != 0 {
		Errorf("call to Setattr with mode irregular")
		return nil
	}

	n.Mode = req.Mode
	return nil
}

// fs.NodeRemover
func (n *fuseFSNode) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	for i, node := range n.Nodes {
		if node.Name == req.Name {
			// TODO: Test if rmdir fills req.Dir
			if req.Dir {
				if !node.Mode.IsDir() {
					return syscall.ENOTDIR
				}
				if len(req.Name) != 0 && req.Name[len(req.Name)-1] == '.' {
					return syscall.EINVAL
				}
			} else {
				if node.Mode.IsDir() {
					return syscall.EISDIR
				}
			}

			n.Nodes = append(n.Nodes[:i], n.Nodes[i+1:]...)
			return nil
		}
	}
	return syscall.ENOENT
}

// fs.NodeStringLookuper
func (n fuseFSNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, n := range n.Nodes {
		if n.Name == name {
			return n, nil
		} else if n.Mode.IsDir() {
			// TODO: Check if this is needed
			if lookupNode, err := n.Lookup(ctx, name); err == nil {
				return lookupNode, nil
			}
		}
	}
	return nil, syscall.ENOENT
}

// fs.NodeMkdirer
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

// fs.NodeCreater
/*
Unused fields
	type CreateRequest struct {
		Flags  OpenFlags
		Umask os.FileMode
	}

	type CreateResponse struct {
		LookupResponse
		OpenResponse
	}
*/
func (n *fuseFSNode) Create(ctx context.Context, req *fuse.CreateRequest, res *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
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
	return newNode, newNode, nil
}

// fs.NodeGetxattrer
func (n fuseFSNode) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, res *fuse.GetxattrResponse) error {
	// NOTE: req.Size is the size of res.Xattr. Size check is performed by fuse library

	if n.Xattrs == nil {
		return syscall.ENODATA
	}

	value, found := n.Xattrs[req.Name]
	if !found {
		return syscall.ENODATA
	}

	res.Xattr = []byte(value)
	return nil
}

// fs.HandleReadDirAller
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

// fs.HandleWriter
/*
Unused fields
	type WriteRequest struct {
		Handle    HandleID
		Offset    int64
		Flags     WriteFlags
		LockOwner LockOwner
	}
*/
func (n *fuseFSNode) Write(ctx context.Context, req *fuse.WriteRequest, res *fuse.WriteResponse) error {
	if req.FileFlags.IsReadOnly() {
		return syscall.EBADF
	}

	// TODO: Get request GID+UID and file UID+GID and check if user or group is allowed to write to the file. If not return EPERM

	n.Data = req.Data
	res.Size = len(req.Data)

	return nil
}
