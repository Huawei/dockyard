# dockyard

A image hub for rkt &amp; docker and other container engine.

## `runtime.conf` 

```
runmode = dev

listenmode = https
httpscertfile = cert/containerops/containerops.crt
httpskeyfile = cert/containerops/containerops.key


[log]
filepath = log/containerops-log

[db]
uri = localhost:6379
passwd = containerops
db = 8

[backend]
driver = gcs

[qiniu]
endpoint = 7xjdb3.com1.z0.glb.clouddn.com
bucket = dockyard
accessKeyID = hYGzs9HMmco3OHosrZ7l6AWQAY9jZghJ9_NE_YBc
accessKeysecret = hDZmVMabRWmlfDze4jPND3UhE_2ce8R93XERFsY1

[aliyun]
endpoint = oss-cn-shenzhen.aliyuncs.com
bucket = dockyard
accessKeyID = cBzEDM4r1oFbn8Zu
accessKeysecret = mE9hgT1Hy4K2VWZq9Ok36Jk3o1AnPw

[upyun]
endpoint = Auto
bucket = dockyard
usr = silverry
passwd = lgp257029

[gcs]
bucketname = dockyad
jsonkeyfile = ./gcs/key.json
projectid = dockyad-test

```
