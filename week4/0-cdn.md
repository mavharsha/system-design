## CDN

```
Assignment: Toy CDN: HTTP server. Caches server on disk. Proxy every request.
```


CDN transparently sits between your user and origin
- as long as its based on HTTP (mostly)

origin -> https://mavharsha.me
Has a image -> https://mavharsha.me/img/sree.jpg


CDN has its own domain.
For user/enterprise the CDN will be https://mavharsha.mycdn.net


To take advantage of CDN
https://mavharsha.mycdn.net/img/sree.jpg



user -----> CDN -----> origin

If the CDN has it cached, it returns the image
If the CDN doesn't have it cached, then the CDN replace host
Example: 
https://mavharsha.mycdn.net/img/sree.jpg --->  https://mavharsha.me/img/sree.jpg
- Replace mavharsha.mycdn.net with mavharsha.me


