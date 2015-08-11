# dockyard

A image hub for rkt &amp; docker and other container engine.

## `runtime.conf` Example

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

[dockyard]
path = data
domains = containerops.me
registry = 0.9
distribution = registry/2.0
standalone = true
```
