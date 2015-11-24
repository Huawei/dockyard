# Dockyard: An image hub for containers

## What is Dockyard
Dockyard is an image hub for docker, rkt or other container engines.

## How it works

## Why it matters
With dockyard you can manage your container images as freely as you can, you need not to concern with different container engines, and you will not be locked in by docker hub.

## Roadmap

## Try it out
Although dockyard is still in development, we encourage you to try out the tool and give feedback. 

### Build
Installation is as simple as:

```bash
go get github.com/containerops/dockyard
```

or as involved as:

```bash
# create a 'github.com/containerops' directory in your GOPATH/src
cd github.com/containerops
git clone https://github.com/containerops/dockyard
cd dockyard
make
sudo make install
```

### Preliminary work
It is quite easy to use dockyard, only a little work should be done before starting dockyard service. Please follow the instructions as below.

#### Dockyard runtime configuration
Please add a runtime config file named `runtime.conf` under `dockyard/conf` before starting `dockyard` service. Below is a `runtime.conf` example:

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
endpoint = xxx
bucket = xxx
accessKeyID = xxx
accessKeysecret = xxx
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

#### Dockyard middleware configuration
Specify parameters to enable Dockyard notification function. Below is an example of `config.json`

```ini
{
   "notifications":{
      "name":"notifications",
      "endpoints":[
         {
            "name":"notifyProxy",
            "url":"http://notifyproxy:8088/events",
            "headers":{"Authorization":["Bearer","token"]},
            "timeout":5000,
            "threshold":5,
            "backoff":5000,
            "eventdb":"/tmp",
            "disabled":false
         }
      ]
   }
}
```

#### Nginx configuration
It's a Nginx config example. You can change **client_max_body_size** what limited upload file size. You should copy `containerops.me` keys from `cert/containerops.me` to `/etc/nginx`, then run **Dockyard** with `http` mode and listen on `127.0.0.1:9911`.

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

### Start dockyard service
- Run directly:

```bash
./dockyard web --address 0.0.0.0 --port 80
```

- Run behind Nginx:

```bash
./dockyard web --address 127.0.0.1 --port 9911
```

### Enjoy it
Congratulations! Dockyard is ready for you, enjoy it:-)
- Add **containerops.me** in your `hosts` file like `192.168.1.66 containerops.me` with IP which run `dockyard`.
- Then `push` with `docker push containerops.me/somebody/ubuntu`.
- You could `pull` with `docker pull -a containerops.me/somebody/ubuntu`.
- Work fun!

## How to involve
If any issues are encountered while using the dockyard project, several avenues are available for support:
<table>
<tr>
	<th align="left">
	Issue Tracker
	</th>
	<td>
	https://github.com/containerops/dockyard/issues
	</td>
</tr>
<tr>
	<th align="left">
	Google Groups
	</th>
	<td>
	https://groups.google.com/forum/#!forum/dockyard-dev
	</td>
</tr>
</table>


## Who should join
- Ones who want to choose a container image hub instead of docker hub.

## Certificate of Origin
By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. 

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

## Format of the Commit Message

You just add a line to every git commit message, like this:

    Signed-off-by: Meaglith Ma <maquanyi@huawei.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.
