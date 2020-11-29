/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

// go build -buildmode=plugin

package main

import (
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Decoder struct {
	id uint64
}

func (d *Decoder) ID() uint64 {
	return atomic.AddUint64(&d.id, 1)
}

func (d *Decoder) Request(in io.Reader, out io.Writer) {
	out.Write([]byte("decoder.so Request\n"))
	io.Copy(out, in)
}

func (d *Decoder) Response(in io.Reader, out io.Writer) {
	out.Write([]byte("decoder.so Response\n"))
	io.Copy(out, in)
}

func (d *Decoder) Close() error {
	return nil
}

func NewDecoderFunc(conn net.Conn) *Decoder {
	if conn != nil {
		log.Println("NewDecoder for conn:", conn.RemoteAddr())
	}
	return &Decoder{}
}
