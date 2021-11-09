# verb-tamper

### HTTP verb tampering & methods enumeration



## Install
```
go install github.com/mmta41/verb-tamper
```


## Usage
```
verb-tamper -url https://host.com/url/to/tamper
```


## Options

```
  -h value
        request headers ex: -h 'X-Api-Key: MyApiKey' -h 'Content-Type: application/json'
  -json
        Output format as json
  -silent
        Disable banner
  -stdin
        Read targets url from stdin
  -t int
            Seconds to wait before timeout. (default 10)
  -url string
        comma separated Urls to check
```