## Docker Registry V2 Protocol

Offical Docker Registry V2 Doc is [here](https://github.com/docker/distribution/blob/master/docs/spec/api.md).

### Docker Registry V2 Ping

1.1 (Docker Client -> Docker Registry) `GET /v2` 
> https://github.com/docker/distribution/blob/master/docs/spec/api.md#api-version-check

### Docker Registry V2 Push

2.1 (Docker Client -> Docker Registry) `PUT /:namespace/:repository/blobs/uploads`

### Docker Registry V2 Pull 