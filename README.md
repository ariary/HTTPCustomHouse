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

**Show the residue of the request that has not been treated** (in stderr)
```shell
cat samples/cl.te | httpcustomhouse -te -r  # -r (or --residue) works also for -cl
```

## Why ?

## Install
