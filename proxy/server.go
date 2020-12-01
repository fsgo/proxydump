/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package proxy

import (
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	DestAddr string

	OnNewConn func(conn net.Conn) (net.Conn, error)

	OnConnClose func(conn net.Conn)

	RequestDumpWriter io.WriteCloser

	ResponseDumpWriter io.WriteCloser

	NewDecoderFunc NewDecoderFunc
}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Accept error:", err.Error())
			continue
		}
		go s.handler(conn)
	}
}

func (s *Server) handler(inConn net.Conn) {
	var err error
	if s.OnNewConn != nil {
		inConn, err = s.OnNewConn(inConn)
	}

	defer inConn.Close()

	if err != nil {
		return
	}

	if s.OnConnClose != nil {
		defer s.OnConnClose(inConn)
	}

	var decoder Decoder
	if s.NewDecoderFunc == nil {
		decoder = NewNopDecoderFunc(inConn)
	} else {
		decoder = s.NewDecoderFunc(inConn)
	}

	errc := make(chan error, 2)

	copy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	outConn, err := net.DialTimeout("tcp", s.DestAddr, 3*time.Second)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer outConn.Close()

	var wOut io.Writer = outConn
	if s.RequestDumpWriter != nil {
		w1 := NewDecoderWriter(s.RequestDumpWriter, decoder.Request)
		defer w1.Close()
		wOut = io.MultiWriter(outConn, w1)
	}
	go copy(wOut, inConn)

	var wIn io.Writer = inConn
	if s.ResponseDumpWriter != nil {
		w2 := NewDecoderWriter(s.ResponseDumpWriter, decoder.Response)
		defer w2.Close()
		wIn = io.MultiWriter(inConn, w2)
	}
	go copy(wIn, outConn)
	<-errc
}
