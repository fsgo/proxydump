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
	Request(p []byte) []byte

	// 解析response
	Response(p []byte) []byte

	Close() error
}

func NewNopDecoderFunc(_ net.Conn) Decoder {
	return &NopDecoder{}
}

var _ Decoder = (*NopDecoder)(nil)

type NopDecoder struct {
}

func (n NopDecoder) Request(p []byte) []byte {
	return p
}

func (n NopDecoder) Response(p []byte) []byte {
	return p
}

func (n NopDecoder) Close() error {
	return nil
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

func NewDecoderWriter(w io.Writer, decoder func([]byte) []byte) io.Writer {
	return &writer{
		raw:         w,
		decoderFunc: decoder,
	}
}

var _ io.Writer = (*writer)(nil)

type writer struct {
	raw         io.Writer
	decoderFunc func([]byte) []byte
}

func (w writer) Write(p []byte) (n int, err error) {
	p1 := w.decoderFunc(p)
	m, err := w.raw.Write(p1)
	if err != nil {
		return m, err
	}
	return len(p), nil
}
