# Dockyard Volume Management Spec
`Volume` is an important resource to a runtime, decoupling it from 'runtime' is necessary. `Dockyard Volume Management` provides an easy and flexible way for container runtime to use volume resource either on a local/remote machine or a cluster. `Dockyard Volume Managment` defines a volume discovery mechanism and implements it by a [`volume manager`](#volume-manager) and an [`agent`](#agent). Following its specification, volume usage becomes quite easy:
 1. query the volume resource and get a volume list
 2. apply a required volume resource and get a local mount point
 3. use the mount point directly by a runtime, either in the command line or the configuration file
 4. free the used volume when the runtime exist or die.
 
##Volume Manager
'Volume Manager` is used to manage all the volume resources inside a cluster.

The [configuration](#volume-manager-configs "configuration") file is used for the Dockyard amdin to set his/her own configuration.

The [APIs](#volume-manager-apis "APIs") are used for the runtime users to list/apply/remove volumes.


###Volume Manager Configs
|Key|Type|Description|Example|
|------|----|------| ----- |
| Port | int | The port of the Scheduler.| 8001 |
| Debug | bool | Print the debug information on the screen| true, default to false |

```
{
	"Port": 8001,
	"Debug": true
}
```

###APIs
####APIs only for Dockyard Agent
|Method|Path|Summary|Description|
|------|----|------|-----------|
| PUT | `/volume` | [Add a volume](#add-a-volume "Add") | Add a new volume. |
| POST | `/volume/:ID` | [Update a volume](#update-a-volume "Update") | Update an exist volume. |
| DELETE| `/volume/:ID` | [Remove a volume](#remove-a-volume "Remove") | Remove a volume. |

####APIs for both container users and Dockyard Agent
|Method|Path|Summary|Description|
|------|----|------|-----------|
| GET | `/volume` | [List Volumes](#list-volume "List") | List volumes on the cluster. |
| GET | `/volume/:ID` | [Detailed volume](#get-volume-status "Details") | Get the detailed information of a volume.|
| POST | `/volume/:ID/apply` | [Apply a certain volume](#apply-a-certain-volume "Apply") | Apply a certain volume to use.|
| POST | `/volume/apply` | [Apply a free volume](#apply-a-free-volume "Apply") | Apply a volume to use.|
| POST | `/volume/:ID/free` | [Free a used volume](#free-a-volume-resource "Free") | Free a volume resource.|

####Add a volume
Add a volume to the 'Volume' server.
When we have a new agent started, the agent should send its volume infos to the 'Volume' server.

```
PUT /volume
```

**Input**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| host | string | The host of the added volume.|
| device | object | The device information of the volume.|
| capbility | object | The command of the action.|

```
curl -X put  {"host": "192.168.100.1", "device": {"path": "/dev/sda1", "type": "ext4", "paras": "ro"}, "capability": {"whole": "4G", "used": "1G"}}  dockyard-volume-server.io/volume
```

**Response**

If success, return the volume id as the body information.
```
  {"ID": "12345678"}
```

####Update a volume
Update the volume information.

```
POST /volume/:ID
```

**Input**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| capbility | object | The command of the action.|

```
curl -d  {"capability": {"whole": "4G", "used": "3G"}}  dockyard-volume-server.io/volume/12345678
```

**Response**

If success,  return `OK` in the http head.

####Remove a volume
Remove a volume, for example, when the underling device is out of use.

```
DELETE /volume/:ID
```

**Input**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| force | bool | Umount the in used connection anyway, default to `false`.|

```
curl -X delete  {"force": false}  dockyard-volume-server.io/volume/12345678
```

**Response**

If success,  return `OK` in the http head, or fill in the error message.


####List volumes
List volumes, with page/pagesize support

```
GET /volume
```

**Parameters**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| page | int | The page size of the volume list.|
| pagesize | int | The listed volumes in one page. |
| status | string | ["free", "in used"(but can share), "locked"]|

```
curl dockyard-volume-server.io/volume?status=free&page=1&pagesize=2
```

**Response**

```
[
       {
       "ID": "8c0bc6d41ff780c05131a2d98a1c00a9",
       "host": "192.168.100.1",
       "capability": {"whole": "4G", "used": "2G"},
       "device": {"path": "/dev/sda1", "type": "ext4", "paras": "ro"},
       "Status": "free",
       },
       {
       "ID": "31cbc4ff780c05131a2d98a1c34a9",
       "host": "192.168.100.1",
       "capability": {"whole": "10G", "used": "0"},
       "device": {"path": "/dev/sda5", "type": "ext4", "paras": "ro"},
       "Status": "free",
       },
]
```

####Get volume status
Get volume's inforamtion.

```
GET /volume/:ID
```

```
curl dockyard-volume-server.io/volume/12345678
```

**Response**

```
       {
       "ID": "8c0bc6d41ff780c05131a2d98a1c00a9",
       "host": "192.168.100.1",
       "capability": {"whole": "4G", "used": "2G"},
       "device": {"path": "/dev/sda1", "type": "ext4", "paras": "ro"},
       "Status": "free",
       }
```

####Apply a free volume
Apply a free volume

```
POST /volume/apply
```

```
curl -d  {}  dockyard-volume-server.io/volume/12345678/apply
```

**Response**

```
{
	"path": /var/lib/1234234434"
}
```

####Apply a certain volume
Apply a certain volume with ID

```
POST /volume/:ID/apply
```

**Input**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| path | string | user defined path. |

```
curl -d  {"path": /data/mypath"}  dockyard-volume-server.io/volume/12345678/apply
```

**Response**

```
{
	"path": /data/mypath"
}
```


####Free a volume
Free a volume, for example when a container stops.

```
POST /volume/:ID/free
```


```
curl -d  {}  dockyard-volume-server.io/volume/12345678/free
```

**Response**

If success,  return `OK` in the http head.



##Agent
'Agent` is used to manage the container related resource in a single node.

The [configuration](#agent-configs "configuration") file is used for the Dockyard admin to set his/her own configuration on a node.

The [APIs](#agent-apis "APIs") are used for the Dockyard admin or other services (like 'Volume') to manage the node resource.

###Agent Configs
|Key|Type|Description|Example|
|------|----|------| ----- |
| Port | int | The port of the Scheduler.| 8002 |
| Debug | bool | Print the debug information on the screen| true, default to false |

```
{
	"Port": 8002,
	"Debug": true
}
```

###Agent APIs
|Method|Path|Summary|Description|
|------|----|------|-----------|
| GET | `/volume` | [List Volumes](#list-volume "List") | List volumes on the node. |
| GET | `/volume/:ID` | [Detailed volume](#get-volume-status "Details") | Get the detailed information of a volume.|
| POST | `/volume/:ID/apply` | [Apply a certain volume](#apply-a-certain-volume "Apply") | Apply a certain volume to use.|
| POST | `/volume/apply` | [Apply a free volume](#apply-a-free-volume "Apply") | Apply a volume to use.|
| POST | `/volume/:ID/free` | [Free a used volume](#free-a-volume-resource "Free") | Free a volume resource.|
| POST | `/volume/gc` | [Garbage collect](#collect-garbage-mounts "Garbage Collect") | Remove the dead mounts.|

####List volumes
List volumes, with page/pagesize support

```
GET /volume
```

**Parameters**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| page | int | The page size of the volume list.|
| pagesize | int | The listed volumes in one page. |
| status | string | ["free", "in used"(but can share), "locked"]|

```
curl dockyard-volume-server.io/volume?status=free&page=1&pagesize=2
```

**Response**

```
[
       {
       "ID": "8c0bc6d41ff780c05131a2d98a1c00a9",
       "host": "192.168.100.1",
       "capability": {"whole": "4G", "used": "2G"},
       "device": {"path": "/dev/sda1", "type": "ext4", "paras": "ro"},
       "Status": "free",
       },
       {
       "ID": "31cbc4ff780c05131a2d98a1c34a9",
       "host": "192.168.100.1",
       "capability": {"whole": "10G", "used": "0"},
       "device": {"path": "/dev/sda5", "type": "ext4", "paras": "ro"},
       "Status": "free",
       },
]
```

####Get volume status
Get volume's inforamtion.

```
GET /volume/:ID
```

```
curl dockyard-volume-server.io/volume/12345678
```

**Response**

```
       {
       "ID": "8c0bc6d41ff780c05131a2d98a1c00a9",
       "host": "192.168.100.1",
       "capability": {"whole": "4G", "used": "2G"},
       "device": {"path": "/dev/sda1", "type": "ext4", "paras": "ro"},
       "Status": "free",
       }
```

####Apply a free volume
Apply a free volume

```
POST /volume/apply
```

```
curl -d  {}  dockyard-volume-server.io/volume/12345678/apply
```

**Response**

```
{
	"path": /var/lib/1234234434"
}
```

####Apply a certain volume
Apply a certain volume with ID

```
POST /volume/:ID/apply
```

**Input**

| *Name* | *Type* | *Description* |
| -------| ------ | --------- |
| path | string | user defined path. |

```
curl -d  {"path": /data/mypath"}  dockyard-volume-server.io/volume/12345678/apply
```

**Response**

```
{
	"path": /data/mypath"
}
```


####Free a volume
Free a volume, for example when a container stops.

```
POST /volume/:ID/free
```

```
curl -d  {}  dockyard-volume-server.io/volume/12345678/free
```

**Response**

If success,  return `OK` in the http head.

####Garbage Collect
In some cases, the container stop without mention that the mounts is out of use.
Add GC support to free all the dead mounts.

```
POST /volume/:ID/gc
```

```
curl -d  {}  dockyard-volume-server.io/volume/12345678/gc
```

**Response**

If success,  return `OK` in the http head.
