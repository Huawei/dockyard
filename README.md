# dockyard

A image hub for rkt &amp; docker and other container engine.

## `runtime.conf` 

```
runmode = dev

listenmode = https
httpscertfile = cert/containerops/containerops.crt
httpskeyfile = cert/containerops/containerops.key

[log]
filepath = log/containerops-log
```