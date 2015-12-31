# OSS: A build-in distributed Object Storage Service for dockyard
![oss-arch](../docs/oss-arch.jpg "Dockyard")

## Architecture of OSS
OSS=Object Storage Service ,which consists of three parts: APIServer , ChunkMaster and Chunkserver

- APIServer is a stateless proxy service, which provides RESTful API to upload/download and manage image files and get Chunkserver info and file id from chunkmaster. APIServer  choose a suitable chunkserver group to storage image according to Chunkserver information independently.

- ChunkMaster  is a  master node designed to maintain chunkserver information and allocate the file id

- Chunkserver is a storage node for performance and space efficiency. It appends single small image file into large files and maintain file index in memory keeping the IO overhead to a minimum. Normally, a Chunkserver group is consist of 3 Chunkservers and each image allocated to a Chunkserver group stored in all the Chunkservers.

besides, OSS needs the collaboration of a metadb, which stored metadata of images in key-value manner.

## Example of configs 

```ini
[oss]
ossmode= allinone
masterhost = 127.0.0.1
masterport = 8099
metahost = 10.229.40.121
metaport = 3306
dbuser = root
dbpasswd = wang
db = speedy1
limitcsnum = 1
connpoolcapacity = 200
servers = 1_127.0.0.1:7657;1_127.0.0.1:7658;1_127.0.0.1:7659
errlogpath = /usr/local/chunkserver/errlog
datapath = /usr/local/chunkserver/data
chunknum = 2
apiport= 80
apihttpsport=443
partsizemb = 4
```