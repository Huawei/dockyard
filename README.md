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
driver = amazons3cloud

[qiniucloud]
endpoint = 7xjdb3.com1.z0.glb.clouddn.com
bucket = dockyard
accessKeyID = hYGzs9HMmco3OHosrZ7l6AWQAY9jZghJ9_NE_YBc
accessKeysecret = hDZmVMabRWmlfDze4jPND3UhE_2ce8R93XERFsY1

[alicloud]
endpoint = oss-cn-shenzhen.aliyuncs.com
bucket = dockyard
accessKeyID = cBzEDM4r1oFbn8Zu
accessKeysecret = mE9hgT1Hy4K2VWZq9Ok36Jk3o1AnPw

[upcloud]
endpoint = v0.api.upyun.com
bucket = dockyard
user = silverry
passwd = 1234567890

[tencentcloud]
endpoint = cosapi.myqcloud.com
accessID = "11000464"
bucket = test
accessKeyID = AKIDBxM1SkbDzdEtLED1KeQhW8HjW5qRu2R5
accessKeysecret = 4ceCa4wNP10c40QPPDgXdfx5MhvuCBWG

[amazons3cloud]
endpoint = s3-ap-southeast-1.amazonaws.com
bucket = dockyards3
accessKeyID = AKIAIJGFXQ2D32O77JXQ
accessKeysecret = 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


[googlecloud]
projectid = dockyad-test
bucket = dockyad
scope = https://www.googleapis.com/auth/devstorage.full_control
privatekey = googlecloud.key
clientemail = 643511510265-1o2mo8fnmbeuvgu2773ga4727pstigtn@developer.gserviceaccount.com


```
