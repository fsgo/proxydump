// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/30

package proxy

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

var _ io.Writer = (*stdOut)(nil)

type stdOut struct {
	index    atomic.Int64
	typeName string
	prefix   string
	x        bool
}

var line = strings.Repeat("-", 80)

var stdoutMux sync.Mutex

func (s *stdOut) Write(p []byte) (n int, err error) {
	stdoutMux.Lock()
	defer stdoutMux.Unlock()

	if s.x {
		const maxLen = 40
		fmt.Fprintf(os.Stdout, "%s[Len=%d] %s\n", s.writePrefix(), len(p), time.Now().Format(time.DateTime+".99999"))
		lineNo := -1
		startIndex := 0
		for len(p) > 0 {
			lineNo++
			end := len(p)
			if len(p) > maxLen {
				end = maxLen
			}

			lineNoStr := color.YellowString("%03d", lineNo)
			indexStr := color.HiGreenString("[%4d - %4d]", startIndex, startIndex+end)
			content := color.CyanString("%q", p[:end])
			fmt.Fprintf(os.Stdout, "%s\t%s %s\n", lineNoStr, indexStr, content)

			fmt.Fprintf(os.Stdout, "\t%v\n", s.byteFmt(p[:end]))
			p = p[end:]
			startIndex += end
		}
	} else {
		fmt.Fprintf(os.Stdout, "%s[Len=%d]\n%c\n\n", s.writePrefix(), len(p), p)
	}
	return len(p), nil
}

func (s *stdOut) byteFmt(bf []byte) []string {
	ss := make([]string, 0, len(bf))
	for _, b := range bf {
		s := fmt.Sprint(b)
		ss = append(ss, fmt.Sprintf("%3s", s))
	}
	return ss
}

func (s *stdOut) writePrefix() string {
	id := color.RedString("%d", s.index.Add(1))
	return fmt.Sprintf("[%s][%s][%s]", s.typeName, id, s.prefix)
}

func (s *stdOut) WriterClone(prefix string) io.Writer {
	return &stdOut{
		typeName: s.typeName,
		prefix:   prefix,
		x:        s.x,
	}
}

type hasWriterClone interface {
	WriterClone(prefix string) io.Writer
}
