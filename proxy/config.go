// Copyright(C) 2020 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2020/11/7

package proxy

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"plugin"
	"reflect"
)

type Config struct {
	// 服务监听地址,如0.0.0.0：8128
	ListenAddr string

	// 原地址，如 www.baidu.com:80
	DestAddr string

	// 请求内容dump输出的文件地址
	RequestDumpPath string

	RequestDumpWriter io.WriteCloser

	// 响应内容dump输出的文件地址
	ResponseDumpPath string

	ResponseDUmpWriter io.WriteCloser

	// 请求和响应解码 .so 的文件地址
	DecoderPluginPath string

	NewDecoderFunc NewDecoderFunc

	AuthToken string

	// 鉴权方法
	Allow func(conn net.Conn) error
}

func (c *Config) Parser() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen addr is empty")
	}
	if c.DestAddr == "" {
		return fmt.Errorf("remote dest addr is empty")
	}

	if err := c.loadDumpFiles(); err != nil {
		return err
	}

	if err := c.loadDecoder(); err != nil {
		return err
	}
	if c.Allow == nil {
		c.Allow = func(conn net.Conn) error {
			return nil
		}
	}

	return nil
}

func (c *Config) loadDumpFiles() error {
	{
		name := c.RequestDumpPath
		if name == "" {
			// pass 不输出
		} else if name == "-" {
			c.RequestDumpWriter = os.Stdout
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			c.RequestDumpWriter = f
		}
	}

	{
		name := c.ResponseDumpPath
		if name == "" {
			// pass  不输出
		} else if name == c.RequestDumpPath {
			c.ResponseDUmpWriter = c.RequestDumpWriter
		} else if name == "-" {
			c.ResponseDUmpWriter = os.Stdout
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			c.ResponseDUmpWriter = f
		}
	}
	return nil
}

// 从plugin 文件加载 decoder
func (c *Config) loadDecoder() error {
	if c.DecoderPluginPath == "" {
		return nil
	}
	_, err := os.Stat(c.DecoderPluginPath)
	if os.IsNotExist(err) {
		return nil
	}

	p, err := plugin.Open(c.DecoderPluginPath)
	if err != nil {
		return err
	}
	v, err := p.Lookup("NewDecoderFunc")
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	connS, connC := net.Pipe()
	connS.Close()
	defer connC.Close()

	vv := rv.Call([]reflect.Value{reflect.ValueOf(connC)})
	if len(vv) != 1 {
		return fmt.Errorf("new decode with wrong result")
	}
	_, ok := vv[0].Interface().(Decoder)
	if !ok {
		return fmt.Errorf("func NewDecoderFunc's type [%T] not func(net.conn)Deocder", v)
	}
	c.NewDecoderFunc = newDecoder(rv)
	return nil
}

func NewConfigByFlag() *Config {
	var config = &Config{}
	flag.StringVar(&config.ListenAddr, "l", "0.0.0.0:8128", "proxy listen addr")
	flag.StringVar(&config.DestAddr, "dest", "", `remote dest server addr (eg "10.10.1.8:80")`)
	flag.StringVar(&config.RequestDumpPath, "req", "-", "dump request to")
	flag.StringVar(&config.ResponseDumpPath, "resp", "-", "dump response to")
	flag.StringVar(&config.DecoderPluginPath, "decoder", "~/proxydump/decoder.so", "decoder plugin so file path")
	flag.StringVar(&config.AuthToken, "auth", "", "auth token, if not empty, server need auth")

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(out, "usage: %s [flags] [path ...]\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(out, "https://github.com/fsgo/proxydump\n")
		fmt.Fprintf(out, "version: %s\n", Version)
	}
	return config
}
