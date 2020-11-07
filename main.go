/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package main

import (
	"flag"
	"log"
	"net"

	"github.com/fsgo/proxydump/proxy"
)

var config = &proxy.Config{}

func init() {
	flag.StringVar(&config.ListenAddr, "l", "0.0.0.0:8080", "Listen Addr")
	flag.StringVar(&config.DestAddr, "dest", "", `remote dest server addr (eg "10.10.1.8:80")`)
	flag.StringVar(&config.RequestDumpPath, "req_dump", "stdout", "")
	flag.StringVar(&config.ResponseDumpPath, "resp_dump", "stdout", "")
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[proxydump] ")

	if err := config.Check(); err != nil {
		log.Fatalln("config error: ", err.Error())
	}

	l, err := net.Listen("tcp", config.ListenAddr)

	if err != nil {
		log.Fatalln("Listen ", config.ListenAddr, " error: ", err.Error())
	}

	log.Println("proxy listen at:", config.ListenAddr)

	s := &proxy.Server{
		Cf: config,
		OnNewConn: func(conn net.Conn) net.Conn {
			log.Println("conn", conn.RemoteAddr(), "open")
			return conn
		},
		OnConnClose: func(conn net.Conn) {
			log.Println("conn", conn.RemoteAddr(), "closed")
		},
	}

	log.Fatalln("exit: ", s.Serve(l))
}
