# Dockyard 内置对象存储服务(Object Storage Service，OSS)部署指南

## 简介
![oss-arch](../docs/oss-arch.jpg "Dockyard")


Dockyard OSS提供两种部署模式：`allinone`模式与`distribute`模式
> allinone 模式为集中部署模式，将OSS的APIserver、chunkmaster及chunkserver全部部署到一台机器上，主要用于试用OSS以及测试
> distribute模式为分布式部署模式，将OSS的APIserver、chunkmaster部署在主节点上，将chunkserver部署到多台从节点上，

## all in one 模式部署方法

（所有操作在同一台机器上执行）

###第一步： 初始化数据库

在mysql客户端中中执行`oss\scripts\`目录下`oss.sql`脚本


###第二步：制定dockyard后台存储为OSS，并开启OSS APIserver开关

修改dockyard的配置文件`conf/runtime.conf`，在`[dockyard]`下增加
```
driver = oss
ossswitch = enable

```

###第三步：配置chunkmaster

在`oss`目录下新建文件`chunkmaster.conf`,  并在文件中添加如下选项

```ini
servermode=          # allinone 部署模式
masterhost =         # master节点IP
masterport =         # master节点端口
metahost =           # mysql服务器IP
metaport =           # mysql服务器端口
dbuser =             # mysql数据库用户名
dbpasswd =           # mysql数据库密码
db =                 # mysql数据库名称，统一为speedy
limitcsnum =         # 最小chunkserver数量
connpoolcapacity =   # 连接池大小
```

示例：

```ini
servermode= allinone              
masterhost = 10.229.40.140        
masterport = 8099                 
metahost = 10.229.40.121          
metaport = 3306                  
dbuser = root                     
dbpasswd = passsword             
db = speedy                       
limitcsnum = 1                    
connpoolcapacity = 200            
```

###第四步：配置chunkserver

在`oss`目录下新建文件`chunkserver.conf`,  并在文件中添加如下选项

```ini
nodenum=3  # 配置node节点数量

#节点1配置
[node1]
groupid=1  					   # 设置group id, 
ip=0.0.0.0                     # allinone模式ip设置为本机
port=9632                      # allinone模式下端口号不能重复
listenmode=https               # http 或 https
datadir=/root/ossdata          # 数据存储路径
errlogdir=/root/osserrlog      # 日志文件存储路径
chunknum=2                     # chunk数量，每个chunk占用2G空间

#节点2配置
[node2]
groupid=1
ip=0.0.0.0
port=9633
listenmode=https
datadir=/root/ossdata
errlogdir=/root/osserrlog
chunknum=2

#节点3配置
[node3]
groupid=1
ip=0.0.0.0
port=9634
listenmode=https
datadir=/root/ossdata
errlogdir=/root/osserrlog
chunknum=2
```

###第五步：运行dockyard



## distribute 模式部署方法


###第一步： 初始化数据库

在chunkmaster节点mysql客户端中中执行`oss\scripts\`目录下`oss.sql`脚本


###第二步：制定dockyard后台存储为OSS，并开启OSS APIserver开关

修改chunkermaster节点上dockyard的配置文件`conf/runtime.conf`，在`[dockyard]`下增加
```
driver = oss
ossswitch = enable

```

###第三步：配置chunkmaster

在所有节点上`oss`目录下新建文件`chunkmaster.conf`,  并在文件中添加如下选项

```ini
servermode=          # distribute 部署模式
masterhost =         # master节点IP
masterport =         # master节点端口
metahost =           # mysql服务器IP
metaport =           # mysql服务器端口
dbuser =             # mysql数据库用户名
dbpasswd =           # mysql数据库密码
db =                 # mysql数据库名称，统一为speedy
limitcsnum =         # 最小chunkserver节点数量
connpoolcapacity =   # 连接池大小
```

示例：

```ini
servermode= distribute
masterhost = 10.229.40.140
masterport = 8099
metahost = 10.229.40.121
metaport = 3306
dbuser = root
dbpasswd = wang
db = speedy1
limitcsnum = 1
connpoolcapacity = 200
```

###第四步：配置chunkserver

在所有节点上`oss`目录下新建文件`chunkserver.conf`,  并在文件中添加如下选项

```ini
nodenum=3  # 配置node节点数量

#节点1配置
[node1]
groupid=1                     # 设置group id，group id表示server分组，同一个分组内所有的的服务器上都会保存上传文件的副本
ip=10.229.40.120              # chunkserver节点ip
port=9632                     # chunkserver节点端口
listenmode=https              # http或https
datadir=/root/ossdata         # 数据存储路径
errlogdir=/root/osserrlog     # 日志文件存储路径
chunknum=2                    # chunk数量，每个chunk占用2G空间

#节点2配置
[node2]
groupid=1
ip=10.229.40.121
port=9632
listenmode=https
datadir=/root/ossdata
errlogdir=/root/osserrlog
chunknum=2

#节点3配置
[node3]
groupid=1
ip=10.229.40.140
port=9632
listenmode=https
datadir=/root/ossdata
errlogdir=/root/osserrlog
chunknum=2

```


###第五步：先运行所有chunkserver节点上的dockyard，然后再运行chunkmaster节点上的dockyard