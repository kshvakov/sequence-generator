Sequence generator
===
[![GoDoc](https://godoc.org/github.com/kshvakov/sequence-generator?status.svg)](https://godoc.org/github.com/kshvakov/sequence-generator/)


Load balancing and ensuring fault-tolerant configuration

[Nginx](http://nginx.org/)

```

upstream sequencer {
	server server_a:8080 max_fails=5 fail_timeout=60s
	server server_b:8080 max_fails=5 fail_timeout=60s
	server server_c:8080 max_fails=5 fail_timeout=60s
	server server_d:8080 max_fails=5 fail_timeout=60s
}

server {

	listen 192.168.10.101:80;

	charset utf-8;
	
	location / {

		proxe_pass http://sequencer;

		proxy_read_timeout 1s;
		proxy_send_timeout 1s;
		proxy_connect_timeout 2s;

		proxy_next_upstream error timeout invalid_header http_502 http_503 http_504;
	}
}


```

Sequence generator 

```

#server_a

./rest-api-server -increment=4 -offset=1

#server_b
./rest-api-server -increment=4 -offset=2

#server_c
./rest-api-server -increment=4 -offset=3

#server_d
./rest-api-server -increment=4 -offset=4


```
