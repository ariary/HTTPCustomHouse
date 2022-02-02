# HTTPCustomHouse

Simple offline simulation of server behavior helping to forge  HTTP smuggling attack

## Usage

`httpcustomhouse` takes as input an HTTP request.

**Show corresponding request treated by a server based on `Content-Length` Header**:  
```shell
cat samples/cl.te | httpcustomhouse -cl
```

**Show corresponding request treated by a server based on chunked encoding**:
```shell
cat samples/cl.te | httpcustomhouse -te
```

**Show the residue of the request that has not been treated** (in stderr) by adding the flag `-r` (or `--residue`) you can see:
```shell
cat samples/cl.te | httpcustomhouse -te -r  #work also for -cl
```

## Why ?

## Install
