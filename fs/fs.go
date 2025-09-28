// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	p "github.com/sbtlocalization/infinity-tools/parser"

	"github.com/spf13/afero"
)

type fileRecord struct {
	FullName     string
	Type         FileType
	BifFile      string
	FileIndex    uint64
	TilesetIndex uint64
	IsTileset    bool
	FileTime     time.Time
	// the following fields are populated when the bif file is read
	FileLength int64
	FileOffset int64
}

func (r *fileRecord) Name() string {
	return r.FullName
}

func (r *fileRecord) Size() int64 {
	return r.FileLength
}

func (r *fileRecord) Mode() os.FileMode {
	return 0o444 // Read-only
}

func (r *fileRecord) ModTime() time.Time {
	return r.FileTime
}

func (r *fileRecord) IsDir() bool {
	return false
}

func (r *fileRecord) Sys() any {
	return nil
}

type dirRecord struct {
	Type  FileType
	Count int64
}

func (d *dirRecord) Name() string {
	return d.Type.String()
}

func (d *dirRecord) Size() int64 {
	return d.Count
}

func (d *dirRecord) Mode() os.FileMode {
	return 0o444 // Read-only
}

func (d *dirRecord) ModTime() time.Time {
	return time.Time{}
}

func (d *dirRecord) IsDir() bool {
	return true
}

func (d *dirRecord) Sys() any {
	return nil
}

type fileCatalog struct {
	byName map[string]*fileRecord
	byType map[FileType]map[string]*fileRecord
	dirs   map[string]*dirRecord
}

func newFileCatalog() *fileCatalog {
	return &fileCatalog{
		byName: make(map[string]*fileRecord),
		byType: make(map[FileType]map[string]*fileRecord),
		dirs:   make(map[string]*dirRecord),
	}
}

type fileEntry struct {
	file     *afero.File
	refCount int
}

type InfinityFs struct {
	KeyFile  string
	filters  []FileType
	catalog  *fileCatalog
	cache    *BifFileCache
	openBifs map[string]*fileEntry
}

func NewInfinityFs(keyFilePath string, filters ...FileType) *InfinityFs {
	fs := afero.NewOsFs()
	keyFile, err := fs.Open(keyFilePath)
	if err != nil {
		log.Panicln("Error opening key file:", err)
		return nil
	}
	defer keyFile.Close()

	key := p.NewKey()
	stream := kaitai.NewStream(keyFile)
	err = key.Read(stream, nil, key)
	if err != nil {
		log.Panicln("Error reading key file:", err)
		return nil
	}

	resources, err := key.ResEntries()
	if err != nil {
		log.Panicln("Error reading key resources:", err)
		return nil
	}

	catalog := newFileCatalog()
	for _, res := range resources {
		recordType := FileTypeFromParserType(res.Type)

		if len(filters) > 0 && !slices.Contains(filters, recordType) {
			continue
		}

		bif, err := res.Locator.BiffFile()
		if err != nil {
			log.Fatalln("Error getting BIF file:", err)
			continue
		}
		bifPath, err := bif.FilePath()
		if err != nil {
			log.Fatalln("Error getting BIF file path:", err)
			continue
		}

		dirFs := afero.NewBasePathFs(fs, filepath.Dir(keyFilePath))
		bifStat, err := dirFs.Stat(bifPath)
		fileTime := time.Time{}
		if err != nil {
			log.Println("Error stating BIF file:", err)
		} else {
			fileTime = bifStat.ModTime()
		}

		record := &fileRecord{
			FullName:     res.Name + "." + recordType.String(),
			FileTime:     fileTime,
			Type:         recordType,
			BifFile:      bifPath,
			FileIndex:    res.Locator.FileIndex,
			TilesetIndex: res.Locator.TilesetIndex,
			IsTileset:    recordType == FileType_TIS,
			FileLength:   -1,
			FileOffset:   -1,
		}

		catalog.byName[record.FullName] = record
		if catalog.byType[record.Type] == nil {
			catalog.byType[record.Type] = make(map[string]*fileRecord)
		}
		catalog.byType[record.Type][record.FullName] = record
	}

	for fileType, records := range catalog.byType {
		catalog.dirs[fileType.String()] = &dirRecord{
			Type:  fileType,
			Count: int64(len(records)),
		}
	}

	cache, err := NewBifFileCache(filepath.Dir(keyFilePath), 10)
	if err != nil {
		log.Panicln("Error creating BIF file cache:", err)
		return nil
	}

	return &InfinityFs{
		KeyFile:  keyFilePath,
		filters:  filters,
		catalog:  catalog,
		cache:    cache,
		openBifs: make(map[string]*fileEntry),
	}
}

// Ensure InfinityFs implements afero.Fs interface
var _ afero.Fs = (*InfinityFs)(nil)

// Create creates a file in the filesystem, returning the file and an error, if any happens.
func (fs *InfinityFs) Create(name string) (afero.File, error) {
	return nil, os.ErrPermission
}

// Mkdir creates a directory in the filesystem, return an error if any happens.
func (fs *InfinityFs) Mkdir(name string, perm os.FileMode) error {
	return os.ErrPermission
}

// MkdirAll creates a directory path and all parents that does not exist yet.
func (fs *InfinityFs) MkdirAll(path string, perm os.FileMode) error {
	return os.ErrPermission
}

// Open opens a file, returning it or an error, if any happens.
func (fs *InfinityFs) Open(name string) (afero.File, error) {
	if FileTypeFromExtension(name) != FileType_Invalid {
		if dir, ok := fs.catalog.dirs[name]; ok {
			return NewInfinityDir(fs, dir), nil
		} else {
			return nil, os.ErrNotExist
		}
	} else {
		return fs.openFile(name)
	}
}

func (fs *InfinityFs) openFile(name string) (afero.File, error) {
	if record, ok := fs.catalog.byName[name]; ok {
		if bifStream, err := fs.openBif(record.BifFile); err == nil {
			bif := p.NewBif()
			stream := kaitai.NewStream(bifStream)
			err = bif.Read(stream, nil, bif)
			if err != nil {
				log.Println("Error reading BIF file", record.BifFile)
				return nil, err
			}

			if !record.IsTileset {
				entries, err := bif.FileEntries()
				if err != nil {
					log.Println("Error reading file entries from", record.BifFile)
					return nil, err
				}
				if int(record.FileIndex) >= len(entries) {
					log.Printf("File index %d out of range in BIF file %s", record.FileIndex, record.BifFile)
					return nil, os.ErrNotExist
				}

				entry := entries[record.FileIndex]
				if entry.Locator.FileIndex != record.FileIndex {
					log.Printf(
						"Warning: File index mismatch in BIF file %s: expected %d, got %d",
						record.BifFile,
						record.FileIndex,
						entry.Locator.FileIndex,
					)
				}

				record.FileLength = int64(entry.LenData)
				record.FileOffset = int64(entry.OfsData)
			} else {
				entries, err := bif.TilesetEntries()
				if err != nil {
					log.Printf("Error reading BIF tileset entries %s", record.BifFile)
					return nil, err
				}
				if int(record.TilesetIndex) >= len(entries) {
					log.Printf("Tileset index %d out of range in BIF file %s", record.TilesetIndex, record.BifFile)
					return nil, os.ErrNotExist
				}

				entry := entries[record.TilesetIndex]
				if entry.Locator.TilesetIndex != record.TilesetIndex {
					log.Printf("Warning: Tileset index mismatch in BIF file %s: expected %d, got %d", record.BifFile, record.TilesetIndex, entry.Locator.TilesetIndex)
				}

				record.FileLength = int64(entry.NumTiles * entry.LenTile)
				record.FileOffset = int64(entry.OfsData)
			}

			bifStream.Seek(0, io.SeekStart) // Reset stream to start
			return NewInfinityFile(fs, record, bifStream), nil
		} else {
			return nil, err
		}
	} else {
		return nil, os.ErrNotExist
	}
}

// OpenFile opens a file using the given flags and the given mode.
func (fs *InfinityFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 || perm != 0o444 {
		return nil, os.ErrPermission
	}
	return fs.Open(name)
}

// Remove removes a file identified by name, returning an error, if any happens.
func (fs *InfinityFs) Remove(name string) error {
	return os.ErrPermission
}

// RemoveAll removes a directory path and any children it contains.
// It does not fail if the path does not exist (return nil).
func (fs *InfinityFs) RemoveAll(path string) error {
	return os.ErrPermission
}

// Rename renames a file.
func (fs *InfinityFs) Rename(oldname, newname string) error {
	return os.ErrPermission
}

// Stat returns a FileInfo describing the named file, or an error, if any happens.
func (fs *InfinityFs) Stat(name string) (os.FileInfo, error) {
	if FileTypeFromExtension(name) != FileType_Invalid {
		if dir, ok := fs.catalog.dirs[name]; ok {
			return dir, nil
		} else {
			return nil, os.ErrNotExist
		}
	} else {
		return fs.statFile(name)
	}
}

func (fs *InfinityFs) statFile(name string) (os.FileInfo, error) {
	if record, ok := fs.catalog.byName[name]; ok {
		if record.FileLength != -1 && record.FileOffset != -1 {
			return record, nil
		}

		if bifFileStream, err := fs.openBif(record.BifFile); err == nil {
			defer fs.closeBif(record.BifFile)

			bif := p.NewBif()
			stream := kaitai.NewStream(bifFileStream)
			err = bif.Read(stream, nil, bif)
			if err != nil {
				log.Println("Error reading BIF file", record.BifFile)
				return nil, err
			}

			if !record.IsTileset {
				entries, err := bif.FileEntries()
				if err != nil {
					log.Println("Error reading file entries from", record.BifFile)
					return nil, err
				}
				if int(record.FileIndex) >= len(entries) {
					log.Printf("File index %d out of range in BIF file %s", record.FileIndex, record.BifFile)
					return nil, os.ErrNotExist
				}
				entry := entries[record.FileIndex]
				record.FileLength = int64(entry.LenData)
				record.FileOffset = int64(entry.OfsData)
				return record, nil
			} else {
				entries, err := bif.TilesetEntries()
				if err != nil {
					log.Println("Error reading tileset entries from", record.BifFile)
					return nil, err
				}
				if int(record.TilesetIndex) >= len(entries) {
					log.Printf("Tileset index %d out of range in BIF file %s", record.TilesetIndex, record.BifFile)
					return nil, os.ErrNotExist
				}
				entry := entries[record.TilesetIndex]
				record.FileLength = int64(entry.NumTiles * entry.LenTile)
				record.FileOffset = int64(entry.OfsData)
				return record, nil
			}
		} else {
			log.Fatalln("Can't open BIF file", record.BifFile)
			return nil, err
		}
	} else {
		return nil, os.ErrNotExist
	}
}

// Name returns the name of this FileSystem.
func (fs *InfinityFs) Name() string {
	return "InfinityFs"
}

// Chmod changes the mode of the named file to mode.
func (fs *InfinityFs) Chmod(name string, mode os.FileMode) error {
	return os.ErrPermission
}

// Chown changes the uid and gid of the named file.
func (fs *InfinityFs) Chown(name string, uid, gid int) error {
	return os.ErrPermission
}

// Chtimes changes the access and modification times of the named file.
func (fs *InfinityFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.ErrPermission
}

func (fs *InfinityFs) openBif(bifPath string) (*io.SectionReader, error) {
	var bifFileEntry *fileEntry

	if entry, ok := fs.openBifs[bifPath]; ok {
		bifFileEntry = entry
	} else if entry, ok := fs.cache.Get(bifPath); ok {
		bifFileEntry = entry
		fs.openBifs[bifPath] = bifFileEntry
	} else {
		log.Fatalln("Can't open BIF file", bifPath)
		return nil, os.ErrClosed
	}

	bifFileEntry.refCount++
	stat, err := (*bifFileEntry.file).Stat()
	if err != nil {
		log.Println("Error stating BIF file", bifPath)
		return nil, err
	}
	section := io.NewSectionReader(*bifFileEntry.file, 0, stat.Size())
	return section, nil
}

func (fs *InfinityFs) closeBif(bifPath string) error {
	if bifFileEntry, ok := fs.openBifs[bifPath]; ok {
		if bifFileEntry.refCount > 0 {
			bifFileEntry.refCount--
			if bifFileEntry.refCount == 0 {
				delete(fs.openBifs, bifPath)
				fs.cache.Add(bifPath, bifFileEntry)
			}
		}
		return nil
	} else {
		log.Fatalln("Can't close BIF file", bifPath)
		return os.ErrClosed
	}
}
