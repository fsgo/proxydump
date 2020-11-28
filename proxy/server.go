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

	OnNewConn func(conn net.Conn) net.Conn

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
	if s.OnNewConn != nil {
		inConn = s.OnNewConn(inConn)
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

	defer inConn.Close()
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
		wOut = io.MultiWriter(outConn, NewDecoderWriter(s.RequestDumpWriter, decoder.Request))
	}
	go copy(wOut, inConn)

	var wIn io.Writer = inConn
	if s.ResponseDumpWriter != nil {
		wIn = io.MultiWriter(inConn, NewDecoderWriter(s.ResponseDumpWriter, decoder.Response))
	}
	go copy(wIn, outConn)
	<-errc
}
