// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/afero"
)

type InfinityFile struct {
	afero.File

	fs     *InfinityFs
	meta   *fileRecord
	stream *io.SectionReader
}

type InfinityDir struct {
	afero.File

	fs   *InfinityFs
	meta *dirRecord
}

func NewInfinityFile(fs *InfinityFs, meta *fileRecord, bifStream *io.SectionReader) *InfinityFile {
	section := io.NewSectionReader(bifStream, meta.FileOffset, meta.FileLength)
	return &InfinityFile{
		fs:     fs,
		meta:   meta,
		stream: section,
	}
}

func (f *InfinityFile) Name() string {
	if f.meta != nil {
		return f.meta.Name()
	} else {
		return ""
	}
}

func (f *InfinityFile) Readdir(count int) ([]os.FileInfo, error) {
	panic("not implemented")
}

func (f *InfinityFile) Readdirnames(n int) ([]string, error) {
	panic("not implemented")
}

func (f *InfinityFile) Stat() (os.FileInfo, error) {
	if f.meta != nil {
		return f.meta, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func (f *InfinityFile) Sync() error {
	return nil
}

func (f *InfinityFile) Truncate(size int64) error {
	return os.ErrPermission
}

func (f *InfinityFile) WriteString(s string) (ret int, err error) {
	return 0, os.ErrPermission
}

func (f *InfinityFile) Close() error {
	if f.fs == nil || f.meta == nil {
		return os.ErrInvalid
	}
	f.stream = nil
	return f.fs.closeBif(f.meta.BifFile)
}

func (f *InfinityFile) Read(p []byte) (n int, err error) {
	if f.stream != nil {
		return f.stream.Read(p)
	} else {
		return 0, os.ErrInvalid
	}
}

func (f *InfinityFile) ReadAt(p []byte, off int64) (n int, err error) {
	if f.stream != nil {
		return f.stream.ReadAt(p, off)
	} else {
		return 0, os.ErrInvalid
	}
}

func (f *InfinityFile) Seek(offset int64, whence int) (int64, error) {
	if f.stream != nil {
		return f.stream.Seek(offset, whence)
	} else {
		return 0, os.ErrInvalid
	}
}

func (f *InfinityFile) Write(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}

func (f *InfinityFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrPermission
}

func NewInfinityDir(fs *InfinityFs, meta *dirRecord) *InfinityDir {
	return &InfinityDir{
		fs:   fs,
		meta: meta,
	}
}

func (d *InfinityDir) Name() string {
	return d.meta.Name()
}

func (d *InfinityDir) Readdir(count int) ([]os.FileInfo, error) {
	if d.fs != nil {
		files := d.fs.catalog.byType[d.meta.Type]
		infos := make([]os.FileInfo, 0, min(len(files), count))
		if count == 0 {
			count = len(files)
		}
		i := 0
		for _, record := range files {
			i++
			if i > count {
				break
			}
			infos = append(infos, record)
		}

		slices.SortFunc(infos, func(a, b os.FileInfo) int {
			return strings.Compare(a.Name(), b.Name())
		})
		return infos, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func (d *InfinityDir) Readdirnames(count int) ([]string, error) {
	if d.fs != nil {
		files := d.fs.catalog.byType[d.meta.Type]
		infos := make([]string, 0, min(len(files), count))
		if count == 0 {
			count = len(files)
		}
		i := 0
		for _, rec := range files {
			i++
			if i > count {
				break
			}
			infos = append(infos, rec.FullName)
		}
		slices.Sort(infos)
		return infos, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func (d *InfinityDir) Stat() (os.FileInfo, error) {
	return d.meta, nil
}

func (d *InfinityDir) Sync() error {
	return nil
}

func (d *InfinityDir) Truncate(size int64) error {
	return os.ErrPermission
}

func (d *InfinityDir) WriteString(s string) (ret int, err error) {
	return 0, os.ErrPermission
}

func (d *InfinityDir) Close() error {
	d.fs = nil
	return nil
}

func (d *InfinityDir) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *InfinityDir) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrInvalid
}

func (d *InfinityDir) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *InfinityDir) Write(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}

func (d *InfinityDir) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, os.ErrPermission
}
