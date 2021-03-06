# miniServer

                _         _  _____                                
               (_)       (_)/  ___|                               
     _ __ ___   _  _ __   _ \ `--.   ___  _ __ __   __  ___  _ __ 
    | '_ ` _ \ | || '_ \ | | `--. \ / _ \| '__|\ \ / / / _ \| '__|
    | | | | | || || | | || |/\__/ /|  __/| |    \ V / |  __/| |   
    |_| |_| |_||_||_| |_||_|\____/  \___||_|     \_/   \___||_|   
                                                              

为当前目录开启一个静态 http 服务器

## 使用教程

### Linux 上面使用
    
    -v 查看版本信息
    
    -p 设定静态服务器运行端口
    
    -k [port number] 杀死特定端口上面的 miniServer
    
    -kl 杀死所有正在运行的 miniServer 程序
    
    -l 查看当前所有正在运行的 miniServer ,返回结果如下
    
           运行端口	        进程id           监听位置
           20001		488		/home/ericwyn/模板/
           20002		462		/home/ericwyn/公共的/
           20003		499		/home/ericwyn/视频

注意:杀死进程依赖于 bash 里 `kill` 命令和 `netstat` 命令

所以你可能需要安装 `net-tools` 

### Windows 上面使用
~~由于 net-tools 在 windows 上面存在使用上的差异， -p , -k , -kl , 功能也无法再 windows 系统上面使用~~

已对windows进行了适配

windows 上面使用 `netstat`, `tasklikt`, `tskill` 等命令来完成功能

基础命令和 linux 上面的完全一样

但是使用 `-l` 的时候无法查看具体的监听文件夹位置


## 日常用法
 - 调试本地的静态网站工程和本地 Javascript( Chrome 是不允许 js 直接通过 ajax 读取本地文件的, 但是使用静态服务器开启网站的时候就可以)

 - Axure PR 将原型图导出后的 HTML 页面，需要安装 Axure PR 的 chrome 插件才能运行，而如果将 miniServer 放到导出的 html 的文件夹目录下，点击运行之后，就可以直接通过 localhost:10010 这样的地址访问了
 
 - 或者是直接开启一个 wifi 分享，内网环境下其他的机器直接通过访问（本机ip地址:端口）的形式来下载文件
 
## 编译
直接 

    go build miniServer.go

就可以了，交叉编译的话在语句前面加上平台限定就好了
 
如果要方便的话就
    
    sudo cp miniServer /usr/bin

这样就直接变成一个 bash 命令了