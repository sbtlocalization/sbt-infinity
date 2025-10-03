// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/spf13/afero"
)

type BifFileCache struct {
	cache    *lru.Cache[string, *fileEntry]
	capacity int
	baseDir  string
}

func NewBifFileCache(baseDir string, capacity int) (*BifFileCache, error) {
	close := func(key string, entry *fileEntry) {
		if entry != nil && entry.refCount <= 0 && entry.file != nil {
			(*entry.file).Close()
		}
	}

	cache, err := lru.NewWithEvict(capacity, close)
	if err != nil {
		return nil, err
	}
	return &BifFileCache{
		cache:    cache,
		capacity: capacity,
		baseDir:  baseDir,
	}, nil
}

func (c *BifFileCache) Add(bifPath string, entry *fileEntry) {
	c.cache.Add(bifPath, entry)
}

func (c *BifFileCache) Get(bifPath string) (*fileEntry, bool) {
	if entry, ok := c.cache.Get(bifPath); ok {
		return entry, true
	}

	fs := afero.NewOsFs()
	dirFs := afero.NewBasePathFs(fs, c.baseDir)
	f, err := dirFs.Open(filepath.FromSlash(bifPath))
	if err != nil {
		return nil, false
	}
	entry := &fileEntry{
		file:     &f,
		refCount: 0,
		parsed:   false,
	}
	c.cache.Add(bifPath, entry)
	return entry, true
}

func (c *BifFileCache) Purge() {
	c.cache.Purge()
}
