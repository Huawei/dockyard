# Updater Server

Dockyard Updater Server is a server side of [software update system](https://github.com/theupdateframework/tuf/#what-is-a-software-update-system)
based on [TUF: The Update Framework](https://www.theupdateframework.com). It receives and provides [appliance](#appliance) from/to a Dockyard Updater Service user.

## Why do we need to adopt TUF to Dockyard
'Securing' Dockyard.
```
The Update Framework (TUF) helps developers to secure new or existing software update systems, which are often found to be vulnerable to many known attacks. TUF addresses this widespread problem by providing a comprehensive, flexible security framework that developers can integrate with any software update system. The framework can be easily integrated (or implemented in the native programming languages of these update systems) due to its concise, self-contained architecture and specification.
```

## What is in its scope
- receives a POST request and sign it automaticly
- provide downloading service for files/meta/signature

## What is NOT in its scope
- authentication

## How to use it
To make a simple demo, you can:
```
	$ mkdir /tmp/dockyard-updater-server-storage -p
        $ cp -fr utils/storage/local/testdata/containerops /tmp/dockyard-updater-server-storage
	$ make
	$ ./dus web
```

### APIs
- list

  ```
	$ curl localhost:1234/app/v1/containerops/official
	{"Message":"AppV1 List files","Content":["appA","appB"]}
  ```

- get file

  ```
	$ curl localhost:1234/app/v1/containerops/official/appA/data > /tmp/appA.txt
	$ cat /tmp/appA.txt
        This is the content of appA.
  ```

- get meta of the whole repository

  ```
	$ curl localhost:1234/app/v1/containerops/official/meta
	{"Message":"AppV1 Get Meta data","Content":[{"Name":"appA","Hash":"4e181e2c1605cfd2b7380afa35e4c6592bb63e21","Created":"2016-07-25T16:49:35.452835973+08:00","Expired":"2017-01-21T16:49:35.452835973+08:00"},{"Name":"appB","Hash":"d89a67deec493af3cb2acc1c7754f7755141ddaa","Created":"2016-07-25T16:49:35.453155584+08:00","Expired":"2017-01-21T16:49:35.453155584+08:00"}]}
  ```

- post file

  ```
	$ curl -d {"this is the content of appX"} localhost:1234/v1/containerops/official/appX
	{"Message":"AppV1 Post data","Content":null}
  ```

### Protocal
  The supported protocal will be `docker/appc/app/image`, now only support `app` (software packages).

### Appliance
  All the docker image, rkt image, software package, vm image are take as an `appliance`.

### ApplianceURL
  Differed with different protocals.
  
  To the `app` protocal, it will be something like 'app://containerops.me/v1/official/dockyard/centos/x86/duc.rpm'.

### Repo URL
  Syntax of repo url should be [:protocal](#protocal)://url:port/[:version](#version)/[:namespace/:repository](#namespace-and-repository)
  ```
	app://containerops.me/v1/official/dockyard

  ```

### Namespace and Repository
Namespace and repository is the same concept in routers.
To a docker user, he/she might pull an image like `library/alpine`,
`library` is the namespace and `alpine` is the repository.

To an app store user, he/she might pull a software like `official/dockyard/centos/x86/duc.rpm`,
`official` is the namespace, `dockyard` is the repository.

### Database
The default location is for a local storage is at "/tmp/dockyard-updater-server-storage"
