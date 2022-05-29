
## captcha-bot

用于[Telegram](https://telegram.org/) 加群验证机器人，采用golang编写，支持全平台编译运行。

<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-blue" alt="license MIT"></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/Golang-1.17-red" alt="Go version 1.17"></a>
<a href="https://github.com/tucnak/telebot"><img src="https://img.shields.io/badge/Telebot Framework-v3-lightgrey" alt="telebot v3"></a>
<a href="https://github.com/assimon/captcha-bot/releases/tag/1.0.0"><img src="https://img.shields.io/badge/version-1.0.0-green" alt="version 1.0.0"></a>
</p>


## 项目初衷
`Telegram`(简称：小飞机)，全球知名的非常方便且优雅的匿名IM工具(比微信更伟大的产品)。    
但由于该软件的匿名性，导致该软件上各种加群推广机器人满天飞，我们无法无时无刻的判断加入群组的“某个人”是否为推广机器人。   
还好`Telegram`为我们提供了非常强大的`Api`，我们可以利用这些Api开发出自动验证的机器人。   

如果你是`Telegram`的群组管理员，你可以直接使用本项目部署私有化的验证机器人。     
如果你是`开发者`，你可以利用本项目熟悉`Go语言`与`Telegram`的交互式开发，以便后续利用`Api`开发出自己的机器人！      

文档参考：   
Telegram Api文档：[Telegram Api](https://core.telegram.org/bots/api)      
机器人开发框架：[Telebot](https://github.com/tucnak/telebot)

## 使用方式

### 一、自行编译
此安装方式多用于开发者，需电脑上安装`go语言`环境。   
[go语言官网](https://golang.org/)    

下载：
```shell
# 下载项目
git clone https://github.com/assimon/captcha-bot && cd captcha-bot && cp .env.example .env
```
编译：
```shell
# 编译
go build -o  cbot
# 给予执行权限
chmod +x ./cbot
```
配置：
```shell
cp .example.config.toml config.toml
```
执行：
```shell
# 调试启动
./cbot
# nohup 
nohup ./cbot >> run.log 2>&1 &
```

### 二、下载已经编译好的二进制程序
此方式可以直接使用，用于服务器生产环境。
进入打包好的版本列表，下载程序：[https://github.com/assimon/captcha-bot/releases](https://github.com/assimon/captcha-bot/releases)    
配置：  
```shell
cp .env.example .env
```
运行：     
```shell
# linux
# 调试启动
./captcha-bot
# nohup 常驻启动

# windows
captcha-bot.exe
```

## 配置：
请将项目目录下`.env.example`文件重命名为`.env`， 然后对`.env`文件进行编辑即可！     
里面的配置项有详细的注释。

## 预览
![禁言.png](https://i.loli.net/2021/09/27/dZQSFKmI23nbXhN.png)
![验证.png](https://i.loli.net/2021/09/27/rEUYVmgt2ve87TL.png)