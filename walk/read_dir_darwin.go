//go:build darwin

package walk

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"syscall"
	"unsafe"
)

func readDir(dirName string) ([]fs.DirEntry, error) {
	var fd uintptr
	var err error
	for {
		fd, err = opendir(dirName)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		return nil, &os.PathError{Op: "opendir", Path: dirName, Err: err}
	}
	defer func() {
		go closedir(fd)
	}()

	var ents []fs.DirEntry

	skipFiles := false
	var dirent syscall.Dirent
	var entptr *syscall.Dirent
	for {
		if errno := readdir_r(fd, &dirent, &entptr); errno != 0 {
			if errno == syscall.EINTR {
				continue
			}
			return nil, &os.PathError{Op: "readdir", Path: dirName, Err: errno}
		}
		if entptr == nil { // EOF
			break
		}
		// Darwin may return a zero inode when a directory entry has been
		// deleted but not yet removed from the directory. The man page for
		// getdirentries(2) states that programs are responsible for skipping
		// those entries:
		//
		//   Users of getdirentries() should skip entries with d_fileno = 0,
		//   as such entries represent files which have been deleted but not
		//   yet removed from the directory entry.
		//
		if dirent.Ino == 0 {
			continue
		}
		typ := dtToType(dirent.Type)
		if skipFiles && typ.IsRegular() {
			continue
		}
		name := (*[len(syscall.Dirent{}.Name)]byte)(unsafe.Pointer(&dirent.Name))[:]
		idx := bytes.IndexByte(name, 0)
		if idx > -1 {
			name = name[:idx]
		}
		// Check for useless names before allocating a string.
		if string(name) == "." || string(name) == ".." {
			continue
		}
		ents = append(ents, &unixDirent{
			name: string(name),
			typ:  typ,
		})
	}

	return ents, nil
}

func dtToType(typ uint8) os.FileMode {
	switch typ {
	case syscall.DT_BLK:
		return os.ModeDevice
	case syscall.DT_CHR:
		return os.ModeDevice | os.ModeCharDevice
	case syscall.DT_DIR:
		return os.ModeDir
	case syscall.DT_FIFO:
		return os.ModeNamedPipe
	case syscall.DT_LNK:
		return os.ModeSymlink
	case syscall.DT_REG:
		return 0
	case syscall.DT_SOCK:
		return os.ModeSocket
	}
	return ^os.FileMode(0)
}

type unixDirent struct {
	name string
	typ  fs.FileMode
}

func (d *unixDirent) Name() string      { return d.name }
func (d *unixDirent) IsDir() bool       { return d.typ.IsDir() }
func (d *unixDirent) Type() fs.FileMode { return d.typ }
func (d *unixDirent) Info() (fs.FileInfo, error) {
	return nil, errors.New("Unsupported")
}
