package fusefs_test

import (
	"context"
	fusefs "fusefs/internal"
	"io/fs"
	"os"
	"syscall"
	"testing"

	"bazil.org/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuseFSNodeWrite(t *testing.T) {
	fileName := "TestFuseFSNodeWrite"
	createdFile, err := os.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer createdFile.Close()
	t.Cleanup(func() {
		os.Remove(fileName)
	})

	t.Run("Write to closed file", func(t *testing.T) {
		// TODO
	})

	t.Run("Write to read-only file", func(t *testing.T) {
		n := fusefs.NewFuseFSNode()
		_, err = n.Open(context.Background(), &fuse.OpenRequest{}, &fuse.OpenResponse{})
		if err != nil {
			t.Fatal(err)
		}
		fuseErrno := n.Write(context.Background(), &fuse.WriteRequest{}, &fuse.WriteResponse{})

		fileRO, err := os.Open(fileName)
		if err != nil {
			t.Fatal(err)
		}
		defer fileRO.Close()

		_, err = fileRO.Write([]byte("test"))
		assert.Error(t, err)
		require.IsType(t, &fs.PathError{}, err)
		pathErr := err.(*fs.PathError)
		errno, ok := pathErr.Err.(syscall.Errno)
		require.True(t, ok)
		assert.Equal(t, errno, fuseErrno)
	})

	t.Run("Write to a directory", func(t *testing.T) {
		// TODO
	})

	t.Run("Success", func(t *testing.T) {
		// TODO
	})
}
