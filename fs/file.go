// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

type InfinityFile struct {
	afero.File

	fs     *InfinityFs
	meta   *fileRecord
	stream *io.SectionReader
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
