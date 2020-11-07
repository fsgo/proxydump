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
	"os"
	"time"
)

type Server struct {
	Cf *Config

	requestFile  *os.File
	responseFile *os.File

	OnNewConn   func(conn net.Conn) net.Conn
	OnConnClose func(conn net.Conn)
}

func (s *Server) init() error {
	if name := s.Cf.RequestDumpPath; name != "" {
		if name == "stdout" {
			s.requestFile = os.Stdout
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			s.requestFile = f
		}
	}

	if name := s.Cf.ResponseDumpPath; name != "" {
		if s.Cf.RequestDumpPath == name {
			s.responseFile = s.requestFile
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			s.responseFile = f
		}
	}
	return nil
}

func (s *Server) Serve(l net.Listener) error {
	s.init()
	for {
		conn, err := l.Accept()
		if err != nil {
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

	defer inConn.Close()
	errc := make(chan error, 2)

	copy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	outConn, err := net.DialTimeout("tcp", s.Cf.DestAddr, 3*time.Second)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer outConn.Close()

	var wOut io.Writer = outConn
	if s.requestFile != nil {
		wOut = io.MultiWriter(outConn, s.requestFile)
	}
	go copy(wOut, inConn)

	var wIn io.Writer = inConn
	if s.responseFile != nil {
		wIn = io.MultiWriter(inConn, s.responseFile)
	}
	go copy(wIn, outConn)
	<-errc
}
