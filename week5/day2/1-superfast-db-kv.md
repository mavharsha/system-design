## Fastest persistant KV store possible

```
Requirements: 

Super fast reads/writes/deletes/full persistance (on HDD)

PUT(k,v)
PUT(k1,v1)
DEL(k)
DEL(k1)
GET(k)
GET(k1)

```


PUT(k,v)
-> appendToFile()
