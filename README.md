# proxydump
代理TCP请求并将传输的内容(请求和响应)Dump。  

HTTP协议是文本协议，直接dump出即可查看。  
若是二进制协议，可以使用单独的程序对Dump的文件进行分析的方案，
也可以自定义Decoder对协议内容进行解析后在Dump。  

Decoder 采用Go Plugin方案，[示例代码详见此](./examples/decoder/decoder.go) 。

## Install
```
go get github.com/fsgo/proxydump
```

## Usage

### 1. 使用`proxydump`启动代理服务
```
proxydump -l "0.0.0.0:8080" -dest "10.10.1.8:80"
```

参数说明：  
`-l "0.0.0.0:8080"` 指定服务监听端口(代理服务器的端口)。  
`-dest "10.10.1.8:80"` 指定目标端口(真实后端服务的地址)。  

默认传输数据会输出到终端(`STDOUT`),也可以输出到指定文件。  


### 2. 修改下游服务配置，将目标ip、port修改为代理服务的监听端口
如 原配置为：
```
Host ：10.10.1.8
Port : 80
```
现在为了抓包分析，修改为：
```
Host ：127.0.0.1
Port : 8080
```

### 3. Example 
以下为抓取访问 http://www.baidu.com/ 的请求数据。  
1.启动代理服务：  
```
proxydump -l '0.0.0.0:8082' -dest "www.baidu.com:80"
```
2.使用curl命令发送请求：
```
curl 'http://127.0.0.1:8082/' -H 'Host: www.baidu.com'
```
3.查看`proxydump`的dump的请求和响应内容：
```
GET / HTTP/1.1
Host: www.baidu.com
User-Agent: curl/7.64.1
Accept: */*

HTTP/1.1 200 OK
Accept-Ranges: bytes
Cache-Control: private, no-cache, no-store, proxy-revalidate, no-transform
Connection: keep-alive
Content-Length: 2381
Content-Type: text/html
Date: Sat, 07 Nov 2020 14:13:42 GMT
Etag: "588604c8-94d"
Last-Modified: Mon, 23 Jan 2017 13:27:36 GMT
Pragma: no-cache
Server: bfe/1.0.8.18
Set-Cookie: BDORZ=27315; max-age=86400; domain=.baidu.com; path=/

// 更多内容略...
```

### 4. IP验证
使用`-auth`参数指定验证token，若IP未验证请求将被拒绝：
```
proxydump -l '0.0.0.0:8082' -dest "www.baidu.com:80" -auth "token_hello"
``` 

启动后，发送验证请求，之后才可以正常使用：
```
echo "token_hello"|nc 127.0.0.1 8082
```