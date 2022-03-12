// Copyright 2022 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tarball

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
)

type MultiTar struct {
	out *gzip.Writer
}

func Out(dst io.Writer) *MultiTar {
	return &MultiTar{
		out: gzip.NewWriter(dst),
	}
}

func (m *MultiTar) Append(ctx *Context, src fs.FS, extra ...io.Writer) error {
	all := make([]io.Writer, 0, len(extra)+1)
	all = append(all, m.out)

	for _, w := range extra {
		all = append(all, gzip.NewWriter(w))
	}

	tw := tar.NewWriter(io.MultiWriter(all...))

	if err := ctx.writeTar(tw, src); err != nil {
		return err
	}

	tw.Flush()

	if len(extra) != 0 {
		// write tar and gzip footers to extra writers to make
		// sure they get a valid archive.
		for _, w := range all[1:] {
			tar.NewWriter(w).Close()
			w.(*gzip.Writer).Close()
		}
	}

	return nil
}

func (m *MultiTar) Close() {
	// write tar and gzip footers to the main writer.
	tar.NewWriter(m.out).Close()
	m.out.Close()
}
