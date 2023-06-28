// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/28

package proxy

import "io"

func cloneWriter(w io.Writer, prefix string) io.Writer {
	if w == nil {
		return w
	}
	if cw, ok := w.(hasWriterClone); ok {
		return cw.WriterClone(prefix)
	}
	return w
}
