# DockerST

本项目旨在于直接加速Docker官方源，提高Docker镜像下载速度
Docker官方使用的是CloudFlare的CDN 优选CloudFlare IP后即可实现高速下载 进行原生加速

## 安装本程序

目前仅在Centos 系统测试过 不过可以确定Linux 系统应该是通用的 Windows 可以用但是没有那么方便

```bash
curl -o /usr/local/bin/DockerST https://github.com/sxhoio/DockerST/releases/download/1.0.0/DockerST_amd64_linux && chmod +x /usr/local/bin/DockerST
```

然后执行 `DockerST` 即可

