package integration_test

import "testing"

// A file is read from where it stands, and putting it somewhere else puts what
// is read after it there, which is what tells a read from a place of its own
// apart from a read that moves along.
func TestFileReadFollowsTheFile(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	path := filepath.Join(os.TempDir(), "go2js_read_follows.txt")
	os.WriteFile(path, []byte("0123456789"), 0o644)

	file, err := os.Open(path)
	fmt.Println("open:", err)

	buffer := make([]byte, 4)

	n, err := file.Read(buffer)
	fmt.Println("read:", n, string(buffer[:n]), err)

	all, err := io.ReadAll(file)
	fmt.Printf("rest: %q %v\n", string(all), err)

	position, err := file.Seek(0, 0)
	fmt.Println("seek:", position, err)

	whole, err := io.ReadAll(file)
	fmt.Printf("whole: %q %v\n", string(whole), err)

	position, err = file.Seek(3, 0)
	fmt.Println("seek:", position, err)

	n, err = file.Read(buffer)
	fmt.Println("read:", n, string(buffer[:n]), err)

	position, err = file.Seek(-2, 0)
	fmt.Println("negative:", position, err)

	file.Close()
	os.Remove(path)
}
`)
}

// A read or a write at a place of its own is carried out there and leaves where
// the file stands alone.
func TestFileReadAtAndWriteAt(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	path := filepath.Join(os.TempDir(), "go2js_at.txt")
	os.WriteFile(path, []byte("0123456789"), 0o644)

	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	fmt.Println("open:", err)

	n, err := file.WriteAt([]byte("AB"), 0)
	fmt.Println("writeat:", n, err)

	buffer := make([]byte, 2)
	n, err = file.ReadAt(buffer, 8)
	fmt.Println("readat:", n, string(buffer), err)

	position, err := file.Seek(0, 0)
	fmt.Println("seek:", position, err)

	head := make([]byte, 2)
	n, err = file.Read(head)
	fmt.Println("read:", n, string(head), err)

	file.Close()

	data, _ := os.ReadFile(path)
	fmt.Println(string(data))

	os.Remove(path)
}
`)
}

// A file that has nothing left reports the end of the text rather than a failure
// of reading it, which is what a read of it to the end answers with.
func TestReadAllStopsAtTheEndOfAFile(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	path := filepath.Join(os.TempDir(), "go2js_read_all.txt")
	os.WriteFile(path, []byte("content"), 0o644)

	file, err := os.Open(path)
	fmt.Println("open:", err)

	all, err := io.ReadAll(file)
	fmt.Printf("%q %v\n", string(all), err)

	empty, err := io.ReadAll(file)
	fmt.Printf("%q %v\n", string(empty), err)

	buffer := make([]byte, 4)
	n, err := file.Read(buffer)
	fmt.Println("read:", n, err)

	file.Close()
	os.Remove(path)
}
`)
}

// What a file says about itself is asked of as methods, which is how a Go
// program asks, and the name it gives is the name in the directory rather than
// the path it was reached by.
func TestFileInformationIsAskedOfAsMethods(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := os.TempDir()
	path := filepath.Join(dir, "go2js_info.txt")
	os.WriteFile(path, []byte("0123456789"), 0o644)

	info, err := os.Stat(path)
	fmt.Println("stat:", err)
	fmt.Println(info.Name(), info.Size(), info.IsDir(), info.Mode() != 0)
	fmt.Printf("%T\n", info)
	fmt.Println(info.ModTime().Year() > 2000)

	file, err := os.Open(path)
	fmt.Println("open:", err)

	own, err := file.Stat()
	fmt.Println("own:", err)
	fmt.Println(own.Name(), own.Size(), own.IsDir())

	file.Close()
	os.Remove(path)

	dirInfo, err := os.Stat(dir)
	fmt.Println("dir:", err, dirInfo.IsDir())
}
`)
}

// A directory is read as the names in it, each asked of as the entry it is.
func TestReadDirGivesEntries(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join(os.TempDir(), "go2js_dir")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)

	entries, err := os.ReadDir(dir)
	fmt.Println("count:", len(entries), err)

	for _, entry := range entries {
		fmt.Println(entry.Name(), entry.IsDir())
	}

	_, err = os.ReadDir(filepath.Join(dir, "no_such_dir"))
	fmt.Println("missing:", err != nil)

	os.RemoveAll(dir)
}
`)
}

// The working directory and the name of the host are answered as a pair, the
// one thing being asked for and whether asking for it went wrong.
func TestWorkingDirectoryAndHostname(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	dir, err := os.Getwd()
	fmt.Println(dir != "", err)

	name, err := os.Hostname()
	fmt.Println(name != "", err)

	environ := os.Environ()
	fmt.Println(len(environ) > 0)

	found := false

	for _, entry := range environ {
		if len(entry) > 0 {
			found = true
		}
	}

	fmt.Println(found)
}
`)
}

// The standard input is a file that is read from, read to the end, put back and
// read again, which is what reading it twice through a scanner needs.
func TestStandardInputIsReadAndPutBack(t *testing.T) {
	runParityTest(t, `package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	all, err := io.ReadAll(os.Stdin)
	fmt.Printf("%d %v\n", len(all), err)

	position, seekErr := os.Stdin.Seek(0, 0)
	fmt.Println("seek:", position, seekErr)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		fmt.Println("line:", scanner.Text())
	}

	fmt.Println("err:", scanner.Err())
}
`)
}

func TestFileModeIsItsOwnType(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// every mark a file mode can carry, written as Go writes them
	modes := []os.FileMode{
		0,
		os.ModeDir,
		os.ModeDir | 0o755,
		0o644,
		0o777,
		os.ModeSymlink | 0o777,
		os.ModeNamedPipe,
		os.ModeSocket,
		os.ModeDevice,
		os.ModeCharDevice | 0o600,
		os.ModeSetuid | 0o755,
		os.ModeSticky | 0o777,
		os.ModeAppend,
		os.ModeExclusive,
		os.ModeTemporary,
		os.ModeIrregular,
	}

	for _, mode := range modes {
		fmt.Println(mode, mode.IsDir(), mode.IsRegular(), mode.Type(), mode.Perm())
	}

	// marks put on a mode leave it a mode rather than a number
	mode := os.FileMode(0o644)
	mode |= 0o111
	mode &^= 0o044

	fmt.Println(mode, mode.IsDir(), mode.Perm())
	fmt.Printf("%v %d %T\n", mode, mode, mode)

	// what a file says about itself is a mode too
	dir := filepath.Join(os.TempDir(), "go2js_modes")
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Join(dir, "inner"), 0o750)
	os.WriteFile(filepath.Join(dir, "file"), []byte("content"), 0o640)

	info, err := os.Stat(dir)
	fmt.Println("dir:", err, info.Mode(), info.Mode().IsDir(), info.Mode().Perm())

	info, err = os.Stat(filepath.Join(dir, "file"))
	fmt.Println("file:", err, info.Mode(), info.Mode().IsRegular(), info.Mode().Perm())

	entries, err := os.ReadDir(dir)
	fmt.Println("entries:", err)

	for _, entry := range entries {
		fmt.Println(entry.Name(), entry.Type(), entry.Type().IsDir(), entry.Type().Perm())
	}

	os.RemoveAll(dir)
}
`)
}
