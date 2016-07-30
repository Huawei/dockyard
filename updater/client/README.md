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
```
	$ make install
	$ man duc
	$ duc help
	NAME:
	   duc - Dockyard Updater client

	USAGE:
	   duc [global options] command [command options] [arguments...]

	VERSION:
	   0.0.1

	COMMANDS:
	    init        initiate default setting
	    add         add a repository url
	    remove      remove a repository url
	    list        list the saved repositories or appliances of a certain repository
	    push	push a file to a repository
	    pull	pull a file from a repository


	GLOBAL OPTIONS:
	   --help, -h           show help
	   --version, -v        print the version
```

### Command lines
- duc init

  Create a default "~/.dockyard/config.json" if it is not exist.
  This is used for a user to set his/her own configurations, see [Config](#config) for details.
  Do not have to 'duc init' if all default settings are already suitable.
- duc add "[repoURL](#repo-url)"

  Add the repo url to the local update list, for example:
  ```
	$ duc add app://localhost:8080/v1/official/dockyard
	"app://localhost:8080/v1/official/dockyard" is added to repo list.
	$ duc add app://localhost:8080/v1/official/dockyard
	"app://localhost:8080/v1/official/dockyard" is already exist in repo list.
  ```
- duc remove "repoURL"

  Remove the `repo url` from the local update list, for example:
  ```
	$ duc remove app://localhost:8080/v1/official/dockyard
	"app://localhost:8080/v1/official/dockyard" is removed from repo list.
  ```
- duc update "repoURL"

  Refresh the `repo`. If "repoURL" is not set, refresh all the local update repo list. For example:
    ```
	$ duc update app://localhost:8080/v1/official/dockyard
	start to refresh "app://localhost:8080/v1/official/dockyard"
	1 new appliances found.
  ```
- duc list "repoURL"
  
  List all the avaliable appliances inside this repo, for example:
  ```
	$ duc list
	app://localhost:8080
	app://containerops.me
	$ duc list app://containerops.me/v1/official/dockyard
	centos/x86/duc.rpm
  ```
- duc pull "filename" "repoULR"
  
  Fetch a certain appliance to [local cache directory](#cache-directory), default to ~/.dockyard/cache.
  No need to set 'repoURL' if 'DefaultServer' is set in ~/.dockyard/config.json.
- duc push "fileURL" "repoURL"
 
  Upload a certain appliance 'fileURL'.
  No need to set 'repoURL' if 'DefaultServer' is set in ~/.dockyard/config.json.

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

### Config
```
   type ducConfig struct {
       DefaultServer string
       CacheDir string
       Repos []string
   }
```

### Cache directory
The default location is '~/.dockyard/Cache', can be reset in '~/.dockyard/config.json'.
