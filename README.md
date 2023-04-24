# drcom-pt
drcom pt 版第三方客户端，使用 Go 编写。

运行后，会每隔一段时间检测网络是否联通，若不则尝试登录。
## 什么是 pt 版
![image](https://user-images.githubusercontent.com/20377726/234033176-5f1b7dc2-aee1-4bbf-b10b-a8678c62af5d.png)  

括号内为 pt 即为 pt 版

## 使用
```
  -a string
        认证地址 (default "http://172.17.100.100:801/eportal/?c=ACSetting&a=Login&jsVersion=3.0&login_t=2")
  -c string
        登录成功后运行的命令
  -p string
        密码
  -u string
        用户名
  -v string
        ver 抓包获取 (default "1.3.5.201712141.P.W.A")
  -z string
        0MKKey 抓包获取 (default "0123456789")
```

认证地址 ver 和 0MKKey 可通过使用 wireshark 之类的抓包工具抓取官方客户端登录时发送的 http 请求获得。

可使用 systemd 运行程序并让其自启。 
