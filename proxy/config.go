/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package proxy

import (
	"fmt"
	"net"
	"os"
	"plugin"
	"reflect"
)

type Config struct {
	ListenAddr string

	DestAddr string

	RequestDumpPath string
	requestFile     *os.File

	ResponseDumpPath string
	responseFile     *os.File

	DecoderPluginPath string
	newDecoderFunc    NewDecoderFunc
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

	return nil
}

func (c *Config) loadDumpFiles() error {
	{
		name := c.RequestDumpPath
		if name == "" || name == "-" {
			// pass
		} else if name == "stdout" {
			c.requestFile = os.Stdout
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			c.requestFile = f
		}
	}

	{
		name := c.ResponseDumpPath
		if name == "" || name == "-" {
			// pass
		} else if name == c.RequestDumpPath {
			c.responseFile = c.requestFile
		} else if name == "stdout" {
			c.responseFile = os.Stdout
		} else {
			f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			c.responseFile = f
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
	c.newDecoderFunc = newDecoder(rv)
	return nil
}

func (c *Config) RequestDumpFile() *os.File {
	return c.requestFile
}

func (c *Config) ResponseDumpFile() *os.File {
	return c.responseFile
}

func (c *Config) NewDecoderFunc() NewDecoderFunc {
	return c.newDecoderFunc
}
