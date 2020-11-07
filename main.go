/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fsgo/proxydump/proxy"
)

var config = &proxy.Config{}

func init() {
	flag.StringVar(&config.ListenAddr, "l", "0.0.0.0:8080", "Listen Addr")
	flag.StringVar(&config.DestAddr, "dest", "", `remote dest server addr (eg "10.10.1.8:80")`)
	flag.StringVar(&config.RequestDumpPath, "req_dump", "stdout", "")
	flag.StringVar(&config.ResponseDumpPath, "resp_dump", "stdout", "")
	flag.StringVar(&config.DecoderPluginPath, "decoder", "~/proxydump/decoder.so", "decoder plugin so file path")

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(out, "usage: %s [flags] [path ...]\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(out, "version: %s\n", proxy.Version)
	}
}

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[proxydump] ")

	if err := config.Parser(); err != nil {
		log.Fatalln("config error: ", err.Error())
	}

	l, err := net.Listen("tcp", config.ListenAddr)

	if err != nil {
		log.Fatalln("Listen ", config.ListenAddr, " error: ", err.Error())
	}

	log.Println("proxy listen at:", config.ListenAddr)

	s := &proxy.Server{
		DestAddr: config.DestAddr,
		OnNewConn: func(conn net.Conn) net.Conn {
			log.Println("conn", conn.RemoteAddr(), "open")
			return conn
		},
		OnConnClose: func(conn net.Conn) {
			log.Println("conn", conn.RemoteAddr(), "closed")
		},
		RequestDumpWriter:  config.RequestDumpFile(),
		ResponseDumpWriter: config.ResponseDumpFile(),
		NewDecoderFunc:     config.NewDecoderFunc(),
	}

	log.Fatalln("exit: ", s.Serve(l))
}
