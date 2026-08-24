# miniServer

```text
              _       _ _____                          
   ____ ___  (_)___  (_) ___/___  ______   _____  _____
  / __ `__ \/ / __ \/ /\__ \/ _ \/ ___/ | / / _ \/ ___/
 / / / / / / / / / / /___/ /  __/ /   | |/ /  __/ /    
/_/ /_/ /_/_/_/ /_/_//____/\___/_/    |___/\___/_/     
                                                       
```

为当前目录或特定目录开启一个静态 http 服务器

简单实现的一个小玩意儿，这些年帮了我很多忙~

当前测试版本：`v3.0.0.beta`（需要 Go `1.22.5` 或更高版本构建）


## 使用教程

### 基本参数
V1.2 版本之后目前所有命令参数都支持 linux / windows

```shell
  -d string
        the dir path the miniServer listen, default is current path
        设置http服务器鉴定目录，默认为当前目录 (default "null")
  -kl
        kill the cross running miniServer
        杀死所有的 miniServer 程序
  -l    list the status of cross running miniServer
        列出当前运行的所有 miniServer 程序
  -p string
        http listen port
        设置 http 服务器运行端口 (default "10010")
  -v    version message
        版本信息查看
```

注意:杀死进程依赖于 bash 里 `kill` 命令和 `netstat` 命令

所以你可能需要安装 `net-tools` 


## 日常用法
 - 调试本地的静态网站工程和本地 Javascript( Chrome 是不允许 js 直接通过 ajax 读取本地文件的, 但是使用静态服务器开启网站的时候就可以)

 - Axure PR 将原型图导出后的 HTML 页面，需要安装 Axure PR 的 chrome 插件才能运行，而如果将 miniServer 放到导出的 html 的文件夹目录下，点击运行之后，就可以直接通过 localhost:10010 这样的地址访问了
 
 - 或者是直接开启一个 wifi 分享，内网环境下其他的机器直接通过访问（本机ip地址:端口）的形式来下载文件

## 下载与断点续传

MiniServer 支持标准 HTTP 字节范围请求，包括：

- `GET` / `HEAD` 请求，以及正确的 `Content-Length`
- 单段和多段 `Range` 请求（`206 Partial Content`）
- `ETag`、`Last-Modified` 和 `If-Range`，用于安全地校验续传文件是否仍是同一版本
- 超出文件范围时返回 `416 Range Not Satisfiable`
- 多个并发分段请求，适用于 aria2 等多线程下载工具
- 大文件和慢速连接；活动下载不会被固定的服务器写入超时中断

例如，使用 8 个连接下载并允许断点续传：

```shell
aria2c -c -x 8 -s 8 -k 1M http://127.0.0.1:10010/path/to/large-file.bin
```

文件名包含空格、`#`、`&` 或中文等字符时，目录页也会生成经过正确 URL 编码的下载链接。

## Web 界面

目录页使用原生 HTML、CSS 和 JavaScript 实现，不依赖框架、CDN、网络字体或其他外部资源。模板资源位于 [`web/`](web/)，并通过 `go:embed` 编译进 MiniServer 二进制文件；部署时仍然只需要复制一个可执行文件。

界面支持桌面与移动端响应式布局、浅色/深色主题、文件搜索、名称/时间/大小/类型排序以及当前目录地址复制。即使关闭 JavaScript，基础目录浏览和文件下载仍然可用。

常见文本、代码、配置和日志文件可以直接在页面内预览。预览通过 HTTP Range 按需读取，最多展示前 `2000` 行或前 `4 MiB`，并提供复制预览全文、跳转到原文件和直接下载操作。

## 编译
直接 

    go build miniServer.go

就可以了，交叉编译的话在语句前面加上平台限定就好了
 
如果要方便的话就
    
    sudo cp miniServer /usr/bin

这样就直接变成一个 bash 命令了
