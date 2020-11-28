/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/11/7
 */

package main

import (
	"flag"
	"log"

	"github.com/fsgo/proxydump/proxy"
)

func main() {
	config := proxy.NewConfigByFlag()
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[proxydump] ")
	log.Fatalln("exit: ", proxy.Run(config))
}
