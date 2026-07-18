// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"testing"

	"github.com/spf13/afero"
)

func TestBuilderCloseWithoutTlk(t *testing.T) {
	b := NewDialogBuilder(afero.NewMemMapFs(), nil, false, false)

	if err := b.Close(); err != nil {
		t.Fatalf("Close on a builder without a TLK file must be a no-op, got %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("Close must be idempotent, got %v", err)
	}
}
