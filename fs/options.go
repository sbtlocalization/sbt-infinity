// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

type fsOptions struct {
	typeFilters   []FileType
	bifFilter     *CompiledFilter
	contentFilter *CompiledFilter
}

// Option is a functional option for NewInfinityFs.
type Option func(*fsOptions)

// WithTypeFilter restricts the catalog to only include resources of the given types.
func WithTypeFilter(types ...FileType) Option {
	return func(o *fsOptions) {
		o.typeFilters = append(o.typeFilters, types...)
	}
}

// WithBifFilter restricts the catalog to resources from BIF files matching
// the given glob pattern. Case insensitive. The "data/" prefix and ".bif"
// extension are stripped before matching unless the pattern contains
// slashes or dots respectively. Empty pattern is a no-op.
func WithBifFilter(pattern string) Option {
	return func(o *fsOptions) {
		o.bifFilter = CompileFilter(pattern, false, true, true)
	}
}

// WithContentFilter restricts the catalog to resources whose name matches
// the given glob pattern. Case insensitive. If the pattern contains no dots,
// the file extension is stripped before matching so bare names like "ABELA01"
// match "ABELA01.WAV". Empty pattern is a no-op.
func WithContentFilter(pattern string) Option {
	return func(o *fsOptions) {
		o.contentFilter = CompileFilter(pattern, false, false, true)
	}
}
