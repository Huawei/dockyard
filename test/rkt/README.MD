#How to test 'rkt'
- [Setup enviornment](#install-the-required-commands)
- [Create testing image](#build-user-image)
- [Modify testconf and Run test](#modify-testsuitconf)

###Install the required commands
```
#Install rkt
wget https://github.com/coreos/rkt/releases/download/v0.15.0/rkt-v0.15.0.tar.gz
tar -zxvf rkt-v0.15.0.tar.gz
cd rkt-v0.15.0
cp rkt /etc/local/bin -f

#Install actools
wget https://github.com/appc/spec/releases/download/v0.7.4/appc-v0.7.4.tar.gz
tar -zxvf appc-v0.7.4.tar.gz
cd appc-v0.7.4
cp actools /etc/local/bin -f

#Install acpush
cd github.com
git clone https://github.com/dgonyeo/acpush.git
cp acpush appc/acpush -r
cd appc/acpush
go build
cp acpush /usr/local/bin/ -f

#Install and setup dockyard
cd $GOPARH/src/github.com
git clone github.com/containerops/dockyard
cd dockyard
make
vim conf/runtime.conf
....
mkdir -p data/acis/pubkeys/user0/etcd
mkdir -p data/acis/pubkeys/user1/etcd
./dockyard web --address 0.0.0.0 --port 443
```

###Build user image
```
cd /home/workspace/aci/acitest/
cd user0  #user0 for example
cat etcd-v2.2.0/manifest
actool build etcd-v2.2.0-linux-amd64/ etcd-v2.2.0-linux-amd64.aci
actool -debug validate etcd-v2.2.0-linux-amd64.aci
cat gpg-batch
gpg --batch --gen-key gpg-batch
gpg --no-default-keyring --armor --secret-keyring ./rkt.sec --keyring ./rkt.pub --output etcd-v2.2.0-linux-amd64.aci.asc  --detach-sig etcd-v2.2.0-linux-amd64.aci
gpg --no-default-keyring --armor --secret-keyring ./rkt.sec --keyring ./rkt.pub --export user0@xxx.com > pubkeys.gpg
cp pubkeys.gpg /home/workspace/gopath/src/github.com/containerops/dockyard/data/acis/pubkeys/user0/etcd/ -f
```

###Modify testsuit.conf
```
domains  = containerops.me
username = user0
pushcmd  = acpush
fetchcmd = rkt
image    = /home/workspace/aci/acitest/etcd-v2.2.0-linux-amd64.aci
```

And now start to test
```
cd test/rkt
go test
```


