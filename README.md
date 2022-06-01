
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
git clone https://github.com/assimon/captcha-bot && cd captcha-bot
```
编译：
```shell
# 编译
go build -o  captcha-bot
# 给予执行权限
chmod +x ./captcha-bot
```
配置：
```shell
cp .example.config.toml config.toml
```
执行：
```shell
# 调试启动
./captcha-bot
# nohup 
nohup ./captcha-bot >> run.log 2>&1 &
```

### 二、下载已经编译好的二进制程序
此方式可以直接使用，用于服务器生产环境。
进入打包好的版本列表，下载程序：[https://github.com/assimon/captcha-bot/releases](https://github.com/assimon/captcha-bot/releases)    
配置：  
```shell
cp .example.config.toml config.toml
```
运行：     
```shell
# linux
# 调试启动
./captcha-bot

# windows
captcha-bot.exe
```

### 三、机器人命令
```
/ping       #存活检测，机器人若正常将返回"pong"
# 广告相关
/add_ad     #新增一条广告，格式：广告标题|跳转链接|到期时间(带时分秒)|权重(倒序，值越大越靠前)，例如：/add_ad 📢广告招租|https://google.com|2099-01-01 00:00:00|100
/all_ad     #查看所有广告
/del_ad     #删除一条广告，例如：/del_ad 1(删除id为1的广告)
```

### 四、敏感词词库使用
在项目`dict`文件夹提供了一些敏感词库，用于机器人反垃圾功能。由于不可描述原因词库不能`明文`放置于项目仓库。      
如需使用，请使用`openssl`命令进行解密，且文件名必须以`dec_`开头，否则无法正常加载！           
例如：     
```shell
openssl enc -d -aes256 -pass pass:captcha-bot -in dict/enc_dc1.txt -out dict/dec_dc1.txt
```

## 预览
![禁言.png](https://i.loli.net/2021/09/27/dZQSFKmI23nbXhN.png)
![验证.png](https://i.loli.net/2021/09/27/rEUYVmgt2ve87TL.png)