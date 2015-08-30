# Dockyard

A image hub for rkt &amp; docker and other container engine.


# How to compile Dockyard application

Make sure Go has been installed and env has been set.

Clone code which Dockyard is depended into directory:

```bash
git clone https://github.com/containerops/dockyard.git $GOPATH/src/github.com/containerops/dockyard
git clone https://github.com/containerops/crew.git $GOPATH/src/github.com/containerops/crew
git clone https://github.com/containerops/wrench.git $GOPATH/src/github.com/containerops/wrench
git clone https://github.com/containerops/ameba.git $GOPATH/src/github.com/containerops/ameba
```

Then exec commands in each project directory as below,it will download the third dependent packages automatically:

```bash
cd $GOPATH/src/github.com/containerops/dockyard
go get

cd $GOPATH/src/github.com/containerops/crew
go get

cd $GOPATH/src/github.com/containerops/wrench
go get

cd $GOPATH/src/github.com/containerops/ameba
go get
```

Finally,enter Dockyard directory and build:
```bash
cd $GOPATH/src/github.com/containerops/dockyard
go build
```


# Dockyard runtime configuration

Please add a runtime config file named `runtime.conf` under `dockyard/conf` before starting `dockyard` service.

## `runtime.conf` Example

```ini
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

[dockyard]
path = data
domains = containerops.me
registry = 0.9
distribution = registry/2.0
standalone = true
driver = qiniu

[qiniu]
access_key = xxx
secret_key = xxx
```

* runmode: application run mode must be `dev` or `prod`.
* listenmode: support `http` and `https` protocol.
* httpscertfile: specify user own https certificate file by this parameter.
* httpskeyfile: specify user own https key file by this parameter.
* [log] filepath: specify where Dockyard logs are stored.
* [db] uri: Dockyard database provider is `redis`,`IP` and `Port` would be specified before `redis` boots.
* [db] passwd: specify the password to login and access db.
* [db] db: specify db area number to use.
* [dockyard] path: specify where `Docker` and `Rocket` image files are stored.
* [dockyard] domains: registry server name or IP.
* [dockyard] registry: specify the version of Docker V1 protocol.
* [dockyard] distribution: specify the version of Docker V2 protocol.
* [dockyard] standalone: must be `true` or `false`,specify run mode whether do authorization checks or not.


# Nginx configuration

It's a Nginx config example. You can change **client_max_body_size** what limited upload file size.

You should copy `containerops.me` keys from `cert/containerops.me` to `/etc/nginx`, then run **Dockyard** with `http` mode and listen on `127.0.0.1:9911`.

```nginx
upstream dockyard_upstream {
  server 127.0.0.1:9911;
}

server {
  listen 80;
  server_name containerops.me;
  rewrite  ^/(.*)$  https://containerops.me/$1  permanent;
}

server {
  listen 443;

  server_name containerops.me;

  access_log /var/log/nginx/containerops-me.log;
  error_log /var/log/nginx/containerops-me-errror.log;

  ssl on;
  ssl_certificate /etc/nginx/containerops.me.crt;
  ssl_certificate_key /etc/nginx/containerops.me.key;

  client_max_body_size 1024m;
  chunked_transfer_encoding on;

  proxy_redirect     off;
  proxy_set_header   X-Real-IP $remote_addr;
  proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_set_header   X-Forwarded-Proto $scheme;
  proxy_set_header   Host $http_host;
  proxy_set_header   X-NginX-Proxy true;
  proxy_set_header   Connection "";
  proxy_http_version 1.1;

  location / {
    proxy_pass         http://dockyard_upstream;
  }
}
```


# How to run

Run directly:

```bash
./dockyard web --address 0.0.0.0 --port 80
```

Run behind Nginx:

```bash
./dockyard web --address 127.0.0.1 --port 9911
```


# How to use

1. Add **containerops.me** in your `hosts` file like `192.168.1.66 containerops.me` with IP which run `dockyard`.
2. Then `push` with `docker push containerops.me/somebody/ubuntu`.
3. You could `pull` with `docker pull -a containerops.me/somebody/ubuntu`.
4. Work fun!


# Reporting issues

Please submit issue at https://github.com/containerops/dockyard/issues


# Maintainers

* Meaglith Ma https://twitter.com/genedna
* Leo Meng https://github.com/fivestarsky


# Licensing

Dockyard is licensed under the MIT License.


# Todo in the feature

1. Support Docker V1/V2 protocol conversion.
2. Support Rocket **CAS**.
3. More relative pages.


# We are working on other projects of Dockyard related

* [Vessel](https://github.com/dockercn/vessel): Continuous Integration Service Core Of ContainerOps.
* [Rudder](https://github.com/dockercn/rudder): Rtk & Docker api client.
