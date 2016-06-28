# Dockyard - Container Registry And Container Repository 
![Dockyard](external/images/dockyard.jpg "Dockyard - Container Registry And Container Volume Mangaement")

## What is Dockyard ?
Dockyard is a container registry storing and distributing container image include [Docker Image](https://github.com/docker/distribution/tree/master/docs/spec), [App Container Image](https://github.com/appc/spec/blob/master/spec/aci.md) and [OCI Image](https://github.com/opencontainer/image-spec). It's key features and goals include:
- Converting image format between above formats.
- Container image encryption, verification and vulnerability analytsis.
- Multi supported distribute protocols include Docker Registry V1 & V2 and App Container Image Discovery.
- Custome distribute protocol by framework base HTTPS and peer to peer. 
- Authentication in distributing process and authorization for public and private container image.
- Supporting mainstream object storage service like Amazon S3, Google Cloud Storage. 
- Built-in object storage service for deployment convenience.
- Volume management with distributed file system and block-based shared storage such as Amazon EBS, OpenStack Cinder.
- Web UI portal for all functions above.

## Why it matters ?

## The Dockyard's Story :)

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
./dockyard web --address 0.0.0.0
```

- Run with Nginx:

```bash
./dockyard web --address 127.0.0.1 --port 9911 &
```

## Update The Libraries Dependencies

```
go get -u -v github.com/Unknwon/com
go get -u -v github.com/aliyun/aliyun-oss-go-sdk/oss
go get -u -v github.com/astaxie/beego
go get -u -v github.com/codegangsta/cli
go get -u -v github.com/docker/libtrust
go get -u -v github.com/go-macaron/inject
go get -u -v github.com/go-sql-driver/mysql
go get -u -v github.com/golang/protobuf/proto
go get -u -v github.com/gorilla/context
go get -u -v github.com/gorilla/mux
go get -u -v github.com/qiniu/api.v6
go get -u -v github.com/qiniu/bytes
go get -u -v github.com/qiniu/rpc
go get -u -v github.com/satori/go.uuid
go get -u -v github.com/upyun/go-sdk/upyun
go get -u -v golang.org/x/crypto/cast5
go get -u -v golang.org/x/crypto/openpgp
go get -u -v golang.org/x/oauth2
go get -u -v google.golang.org/api/gensupport
go get -u -v google.golang.org/api/googleapi
go get -u -v google.golang.org/api/storage/v1
go get -u -v google.golang.org/cloud/compute/metadata
go get -u -v google.golang.org/cloud/internal
go get -u -v gopkg.in/bsm/ratelimit.v1
go get -u -v gopkg.in/ini.v1
go get -u -v gopkg.in/macaron.v1
go get -u -v gopkg.in/redis.v3
go get -u -v github.com/jinzhu/gorm
```

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
- Ones who want to ease the burden of container image management.

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
