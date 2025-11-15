## Word dictionary without any DB

```
Requirements: 
- Given  a word, give the meaning.
- no traditional DBs usage
- 1TB big and 170000 words
- No repetitive enteries
- words and meanings updated weekly --> through change log
- Portable (single file)
- Response time can be high
```

-----

Store word -> meaning

Can use CSV
```csv
word,meaning
-------------
apple,a fruit
ball,an round object
cat,an animal
.
.
.
zoo,home for animals
```

Access

GET /meanings/<word>

Flow:
```
openFile()
readFileLineByLine()
    currentWord,meaning=readLine()
    if currentWord == word:
        return meaning
return null
```

For every word retrieval, full file scan is needed. (1TB)

Need to optimize:
- Maybe create index for the file