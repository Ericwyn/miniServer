# miniServer_GO
用 Go 写的一个小工具，虽然很简单，但是可能对大家有点帮助

## 喵喵喵???
自己平常里经常用到，在某个文件夹下运行的时候，直接就可以把这个文件夹变成静态服务器的根目录，不需要配置 Caddy 之类的，因为本地调试 javaScript 的时候会受到浏览器限制，例如 Chrome 是不允许 js 直接通过 ajax 读取本地文件的。而如果是运行在 localhost:XXXX 上面的网页的话就可以

## 更多
 日常里面发现，例如 AxurePR 将原型图导出后的 HTML 页面，需要安装 AxurePR 的 chrome 插件才能运行，而如果将 miniServer 放到导出的 html 的文件夹目录下，点击运行之后，就可以直接通过 localhost:10010 这样的地址访问了