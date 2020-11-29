/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package proxy

import (
	"fmt"
	"io"
	"net"
	"reflect"
)

type NewDecoderFunc func(conn net.Conn) Decoder

// Decoder 数据解码器
type Decoder interface {
	// 解析请求数据
	Request(in io.Reader, out io.Writer)

	// 解析response
	Response(in io.Reader, out io.Writer)
}

func NewNopDecoderFunc(_ net.Conn) Decoder {
	return &NopDecoder{}
}

var _ Decoder = (*NopDecoder)(nil)

type NopDecoder struct {
}

func (n NopDecoder) Request(in io.Reader, out io.Writer) {
	io.Copy(out, in)
}

func (n NopDecoder) Response(in io.Reader, out io.Writer) {
	io.Copy(out, in)
}

func newDecoder(rv reflect.Value) NewDecoderFunc {
	return func(conn net.Conn) Decoder {
		vv := rv.Call([]reflect.Value{reflect.ValueOf(conn)})
		if len(vv) != 1 {
			panic(fmt.Sprintf("newDecoder with wrong result length,got=%d,want=1", len(vv)))
		}
		decoder, ok := vv[0].Interface().(Decoder)
		if !ok {
			panic("not Decoder")
		}
		return decoder
	}
}

func NewDecoderWriter(out io.Writer, decoder func(in io.Reader, out io.Writer)) io.WriteCloser {
	w := &writer{}
	w.r, w.w = io.Pipe()
	go decoder(w.r, out)
	return w
}

var _ io.Writer = (*writer)(nil)

type writer struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writer) Close() error {
	w.r.CloseWithError(io.EOF)
	w.w.CloseWithError(io.EOF)
	return nil
}
