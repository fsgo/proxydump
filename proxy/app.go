/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/28
 */

package proxy

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
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
		OnNewConn: func(conn net.Conn) (net.Conn, error) {
			log.Println("conn", conn.RemoteAddr(), "open")
			return conn, config.Allow(conn)

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

type authInfo struct {
	Hosts map[string]time.Time
	Token string
	lock  sync.RWMutex
}

func (a *authInfo) ipAllow(ip string) bool {
	a.lock.RLock()
	a.lock.RUnlock()
	_, has := a.Hosts[ip]
	return has
}

func (a *authInfo) Allow(conn net.Conn) error {
	if a.Token == "" {
		return nil
	}
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Printf("SplitHostPort faild %v\n", err)
		return err
	}

	if a.ipAllow(host) {
		return nil
	}

	rd := bufio.NewReader(conn)
	line, _, _ := rd.ReadLine()
	token := string(line)
	if token == a.Token {
		a.lock.Lock()
		defer a.lock.Unlock()
		a.Hosts[host] = time.Now()
		log.Printf("auth success, %q\n", host)
		_, _ = conn.Write([]byte("auth success"))
		return nil
	}
	_, _ = conn.Write([]byte("forbidden"))
	log.Printf("forbidden, %q not auth\n", host)
	return fmt.Errorf("not auth")
}

func WithAuth(config *Config) {
	au := &authInfo{
		Token: config.AuthToken,
		Hosts: map[string]time.Time{},
	}
	config.Allow = func(conn net.Conn) error {
		return au.Allow(conn)
	}
}
