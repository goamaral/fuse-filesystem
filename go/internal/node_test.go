package fusefs

import (
	"io/fs"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GenerateTestFile(t *testing.T, fsPath string) *os.File {
	fileName := strings.ToLower(
		strings.ReplaceAll(
			strings.ReplaceAll(t.Name(), "/", "-"),
			"_", "-",
		),
	)
	file, err := os.CreateTemp(fsPath, fileName)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, file.Close())
		require.NoError(t, os.Remove(file.Name()))
	})

	return file
}

func TestNode_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		generatedFile := GenerateTestFile(t, "/tmp/fusefs")
		str := "hello"
		n, err := generatedFile.WriteString(str)
		assert.Equal(t, len(str), n)
		require.NoError(t, err)
	})

	t.Run("FileIsReadOnly", func(t *testing.T) {
		generatedFile := GenerateTestFile(t, "/tmp/fusefs")
		file, err := os.OpenFile(generatedFile.Name(), os.O_RDONLY, 0)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, file.Close()) })

		n, err := file.WriteString("hello")
		assert.Zero(t, n)
		require.Error(t, err)
		pathErr, ok := err.(*fs.PathError)
		require.True(t, ok, "err is not *fs.PathError")
		errno, ok := pathErr.Err.(syscall.Errno)
		require.True(t, ok, "err is not syscall.Errno")
		assert.Equal(t, syscall.EBADF, errno)
	})
}
