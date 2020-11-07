/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package proxy

import (
	"fmt"
)

type Config struct {
	ListenAddr string

	DestAddr string

	RequestDumpPath string

	ResponseDumpPath string
}

func (c *Config) Check() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen addr is empty")
	}
	if c.DestAddr == "" {
		return fmt.Errorf("remote dest addr is empty")
	}
	return nil
}
