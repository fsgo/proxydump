/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/28
 */

package proxy

import (
	"log"
	"net"
)

func Run(config *Config) error {
	if err := config.Parser(); err != nil {
		return err
	}

	l, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		return err
	}
	log.Println("proxy listen at:", config.ListenAddr)

	s := &Server{
		DestAddr: config.DestAddr,
		OnNewConn: func(conn net.Conn) net.Conn {
			log.Println("conn", conn.RemoteAddr(), "open")
			return conn
		},
		OnConnClose: func(conn net.Conn) {
			log.Println("conn", conn.RemoteAddr(), "closed")
		},
		RequestDumpWriter:  config.RequestDumpWriter,
		ResponseDumpWriter: config.ResponseDUmpWriter,
		NewDecoderFunc:     config.NewDecoderFunc,
	}
	return s.Serve(l)
}
