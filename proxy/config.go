// Copyright(C) 2020 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2020/11/7

package proxy

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"plugin"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/fsgo/fsgo/fsfs"
)

type Config struct {
	// 服务监听地址,如0.0.0.0：8128
	ListenAddr string

	// 原地址，如 www.baidu.com:80
	DestAddr string

	// 请求内容dump输出的文件地址
	RequestDumpPath string

	RequestDumpWriter io.Writer

	// 响应内容dump输出的文件地址
	ResponseDumpPath string

	ResponseDumpWriter io.Writer

	// 最多保留文件数
	MaxFiles int

	// 请求和响应解码 .so 的文件地址
	DecoderPluginPath string

	NewDecoderFunc NewDecoderFunc

	AuthToken string

	// 鉴权方法
	Allow func(conn net.Conn) error

	XBytes bool
}

func (c *Config) Parser() error {
	if len(c.ListenAddr) == 0 {
		return errors.New("listen addr is empty")
	}
	if len(c.DestAddr) == 0 {
		return errors.New("remote dest addr is empty")
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
		if len(name) == 0 || name == "no" {
			// pass 不输出
		} else if name == "stdout" {
			c.RequestDumpWriter = c.stdout(color.GreenString("Request"))
		} else {
			rf, err := c.openFile(name)
			if err != nil {
				return err
			}
			c.RequestDumpWriter = rf
		}
	}

	{
		name := c.ResponseDumpPath
		if len(name) == 0 || name == "no" {
			// pass  不输出
		} else if name == "stdout" {
			c.ResponseDumpWriter = c.stdout(color.YellowString("Response"))
		} else if name == c.RequestDumpPath {
			c.ResponseDumpWriter = c.RequestDumpWriter
		} else {
			rf, err := c.openFile(name)
			if err != nil {
				return err
			}
			c.ResponseDumpWriter = rf
		}
	}
	return nil
}

func (c *Config) stdout(name string) io.Writer {
	return &stdOut{
		typeName: name,
		x:        c.XBytes,
	}
}

var _ io.Writer = (*stdOut)(nil)

type stdOut struct {
	index    atomic.Int64
	typeName string
	prefix   string
	x        bool
}

var line = strings.Repeat("-", 80)

func (s *stdOut) Write(p []byte) (n int, err error) {
	if s.x {
		vs := fmt.Sprint(p)
		fmt.Fprintf(os.Stdout, "%s[Len=%d]\n%c\n%s\n%s\n\n", s.writePrefix(), len(p), p, line, vs)
	} else {
		fmt.Fprintf(os.Stdout, "%s[Len=%d]\n%c\n\n", s.writePrefix(), len(p), p)
	}
	return len(p), nil
}

func (s *stdOut) writePrefix() string {
	id := color.RedString("%d", s.index.Add(1))
	return fmt.Sprintf("[%s][%s][%s]", s.typeName, id, s.prefix)
}

func (s *stdOut) WriterClone(prefix string) io.Writer {
	return &stdOut{
		typeName: s.typeName,
		prefix:   prefix,
		x:        s.x,
	}
}

type hasWriterClone interface {
	WriterClone(prefix string) io.Writer
}

func (c *Config) openFile(name string) (io.WriteCloser, error) {
	maxFiles := c.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 12
	}
	rf := &fsfs.Rotator{
		Path:     name,
		ExtRule:  "1hour",
		MaxFiles: maxFiles,
	}
	if err := rf.Init(); err != nil {
		return nil, err
	}
	return rf, nil
}

// 从plugin 文件加载 decoder
func (c *Config) loadDecoder() error {
	if len(c.DecoderPluginPath) == 0 {
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
		return errors.New("new decode with wrong result")
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
	flag.StringVar(&config.RequestDumpPath, "req", "stdout", `dump request to file
stdout -> os.Stdout
no     -> no output
other case as filepath，eg request.txt、dump/request.txt
`)
	flag.StringVar(&config.ResponseDumpPath, "resp", "stdout", `dump response to file
stdout -> os.Stdout
no     -> no output
other case as filepath，eg response.txt、dump/response.txt
`)
	flag.StringVar(&config.DecoderPluginPath, "decoder", "~/proxydump/decoder.so", "decoder plugin so file path")
	flag.StringVar(&config.AuthToken, "auth", "", "auth token, if not empty, server need auth")
	flag.IntVar(&config.MaxFiles, "max_files", 12, "max dump files to keep")
	flag.BoolVar(&config.XBytes, "x", false, "stdout bytes")

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(out, "usage: %s [flags] [path ...]\n", cmd)
		flag.PrintDefaults()
		fmt.Fprint(out, "https://github.com/fsgo/proxydump\n")
		fmt.Fprintf(out, "version: %s\n", Version)
	}
	return config
}
