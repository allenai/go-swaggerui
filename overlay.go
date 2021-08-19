package swaggerui

import (
	"bytes"
	"io/fs"
)

// overlayFS overlays new files on top of a base FS.
type overlayFS struct {
	fs.FS
	replaced map[string]overlayFile
}

func (fs overlayFS) Open(name string) (fs.File, error) {
	if f, ok := fs.replaced[name]; ok {
		overlaid := f // Make a copy so readers don't modify the original.
		return &overlaid, nil
	}
	return fs.FS.Open(name)
}

// overlayFile masquerades as an existing file with new data.
type overlayFile struct {
	fs.FileInfo
	content bytes.Buffer
}

func (f *overlayFile) Stat() (fs.FileInfo, error) { return f, nil }
func (f *overlayFile) Read(p []byte) (int, error) { return f.content.Read(p) }
func (f *overlayFile) Close() error               { return nil }

func (f *overlayFile) Size() int64 { return int64(f.content.Len()) }
