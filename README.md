Bott
======

A dns service with api

Api
=====

1. list hosts

```
http 127.0.0.1:8080

HTTP/1.1 200 OK
Content-Length: 65
Content-Type: application/json
Date: Sun, 17 Aug 2014 02:45:07 GMT
Server: Bott DNS server

{
    "y.intra.hunantv.com": [
        "10.1.201.1",
        "10.1.201.2",
        "10.1.201.3"
    ]
}
```

2. Bind ips to a host

```
http -j PUT 127.0.0.1:8080/host/y.intra.hunantv.com ip:='["a", "b", "c"]'
```

3. Remove ips to a host

```
http -j DELETE 127.0.0.1:8080/host/y.intra.hunantv.com ip:='["b", "c"]'
```

