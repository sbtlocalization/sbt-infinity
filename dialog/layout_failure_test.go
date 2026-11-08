// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sbtlocalization/dcanvas"
)

func TestLayoutFailureWarnsAndPreservesPositions(t *testing.T) {
	c := &dcanvas.Canvas{Nodes: []*dcanvas.Node{{ID: "a", Type: "text", X: 5, Y: 7, Width: 400, Height: 300}}, Edges: []*dcanvas.Edge{}}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	applyLayout(c, NewNodeOrigin("TEST", 0), func(c *dcanvas.Canvas) error { c.Nodes[0].X = 999; return errors.New("test failure") })
	writer.Close()
	warning, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(warning), "warning(TEST[0]): layout algorithm failed: test failure") {
		t.Errorf("warning: %q", warning)
	}
	if c.Nodes[0].X != 5 || c.Nodes[0].Y != 7 {
		t.Errorf("positions: %+v", c.Nodes[0])
	}
	var out bytes.Buffer
	if err := dcanvas.Encode(c, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := dcanvas.Decode(&out); err != nil {
		t.Fatal(err)
	}
}
