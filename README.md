# HTTPCustomHouse

Simple offline simulation of server behavior helping to forge  HTTP smuggling attack

## Usage

`httpcustomhouse` takes as input an HTTP request.

You can simulate request treatment for a server based on `Content-Length` Header:  
```shell
cat samples/cl.te | httpcustomhouse -cl
```
It will output the **corresponding request treated**

Alternatively, you can simulate request treatment for a server based on chunked encoding:
```shell
cat samples/cl.te | httpcustomhouse -te
```
It will output the **corresponding request treated**

By adding the flag `-r` (or `--residue`) you can see **the residue of the request that has not been treated.** For example, to see the part of my request that hasn't been treated by a server using chunked encoding:
```shell
cat samples/cl.te | httpcustomhouse -te -r
```

## Why ?

## Install
