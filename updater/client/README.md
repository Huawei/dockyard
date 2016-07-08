# Updater Client

Dockyard Updater Client is a [software update system](https://github.com/theupdateframework/tuf/#what-is-a-software-update-system)
based on [TUF: The Update Framework](https://www.theupdateframework.com). It pulls/pushes required [appliance](#appliance) from/to a dockyard server.

## Why do we need to adopt TUF to Dockyard
'Securing' Dockyard.
```
The Update Framework (TUF) helps developers to secure new or existing software update systems, which are often found to be vulnerable to many known attacks. TUF addresses this widespread problem by providing a comprehensive, flexible security framework that developers can integrate with any software update system. The framework can be easily integrated (or implemented in the native programming languages of these update systems) due to its concise, self-contained architecture and specification.
```

## What is in its scope
- bookmark a repo and list its appliance
- download a certain appliance
- verify if a downloaded appliance is valid
- signature and push your appliance to a remote dockyard server

## What is NOT in its scope
- install or run an appliance

## How to use it
- dyclient init

  Create a default "~/.dyclient/config.json" if it is not exist.
  This is used for a user to set his/her own configurations, see [Config](#config) for details.
  Do not have to 'dyclient init' if all default settings are already suitable.
- dyclient add "[repoURL](#repo-url)"

  Add the repo url to the local update list, for example:
  ```
	$ dyclient add app://localhost:8080/v1/dliang/dockyard
	"app://localhost:8080/v1/dliang/dockyard" is added to repo list.
	$ dyclient add app://localhost:8080/v1/dliang/dockyard
	"app://localhost:8080/v1/dliang/dockyard" is already exist in repo list.
  ```
- dyclient remove "repoURL"

  Remove the `repo url` from the local update list, for example:
  ```
	$ dyclient remove app://localhost:8080/v1/dliang/dockyard
	"app://localhost:8080/v1/dliang/dockyard" is removed from repo list.
  ```
- dyclient update "repoURL"

  Refresh the `repo`. If "repoURL" is not set, refresh all the local update repo list. For example:
    ```
	$ dyclient update app://localhost:8080/v1/dliang/dockyard
	start to refresh "app://localhost:8080/v1/dliang/dockyard"
	1 new appliances found.
  ```
- dyclient list "repoURL"
  
  List all the avaliable appliances inside this repo, for example:
  ```
	$ dyclient list app://localhost:8080/v1/dliang/dockyard
	centos/x86/dyclient.rpm
  ```
- dyclient pull --protocal "protocal:version" --server "localhost:8080" --path "localDir" "repoULR"
  
  Fetch a certain appliance to [local cache directory](#cache-directory) if "localDir" is not set.
  No need to set '--server' if 'DefaultServer' is set in ~/.dyclient/config.json.
- dyclient push --protocal "protocal:version" --server "localhost:8080" --path "localDir" "repoURL"
 
  Upload a certain appliance from from local cache directory if "localDir" is not set.
  No need to set '--server' if 'DefaultServer' is set in ~/.dyclient/config.json.

### Protocal
  The supported protocal will be `docker/appc/app/image`, now only support `app` (software packages).

### Appliance
  All the docker image, rkt image, software package, vm image are take as an `appliance`.

### ApplianceURL
  Differed with different protocals.
  
  To the `app` protocal, it will be something like 'app://localhost:8080/v1/dliang/dockyard/centos/x86/dyclient.rpm'.

### Repo URL
  Syntax of repo url should be [:protocal](#protocal)://url:port/[:version](#version)/[:namespace/:repository](#namespace-and-repository)
  ```
	app://localhost:8080/v1/dliang/dockyard

  ```

### Namespace and Repository
Namespace and repository is the same concept in routers.
To a docker user, he/she might pull an image like `library/alpine`,
`library` is the namespace and `alpine` is the repository.

To an app store user, he/she might pull a software like `dliang/dockyard/centos/x86/dyclient.rpm`,
`dliang` is the namespace, `dockyard` is the repository.

### Config
```
   type dyclientConfig struct {
       DefaultServer string
       CacheDir string
   }
```

### Cache directory
The default location is '~/.dyclient/Cache', can be reset in '~/.dyclient/config.json'.
