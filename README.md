# HTTPCustomHouse

<div align=center>
<img src= https://github.com/ariary/HTTPCustomHouse/blob/main/img/E0D8F573-7824-42C1-BF6B-F58E5F14DB0E.png width=150>

<br><strong><i>Simple offline simulation of server behavior helping to forge  HTTP smuggling attack</i></strong>
    
<b>httpcustomhouse</b> <strong>•</strong> analyze smuggle request  
    
<b>httpoverride </b><strong>•</strong>  manipulate raw HTTP request

<b>httpclient</b> <strong>•</strong>  send raw HTTP request

<!---<table>
    <thead>
        <tr>
          <th colspan="2">⬇️ <code>TE.CL</code> example</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td><img src=https://github.com/ariary/HTTPCustomHouse/blob/main/img/hch.png></td>
        </tr>
    </tbody>
</table>-->
</div> 
  
## Usage

`httpcustomhouse` takes as input a raw HTTP request.

**Show corresponding request treated by a server based on `Content-Length` Header**:  
```shell
cat samples/te.cl | httpcustomhouse -cl
```

**Show corresponding request treated by a server based on chunk encoding**:
```shell
cat samples/cl.te | httpcustomhouse -te
```

**Show the residue of the request that has not been treated** (in stderr):
```shell
cat samples/cl.te | httpcustomhouse -te -r
# -r (or --residue) works also for -cl
```

* [Forge `TE.CL` request smuggling attack](#analyze-tecl-request-treatment)
* [Forge `CL.TE` request smuggling attack](#analyze-clte-request-treatment)
* [How to build raw HTTP request?](#building-http-request)
* [How to send raw HTTP request?](#send-raw-http-request)
* [Visualize `TE.CL` (🖼️)](https://github.com/ariary/HTTPCustomHouse/blob/main/img/hch.png)
* [Exploit `CL.TE` (📝)](https://github.com/ariary/HTTPCustomHouse/blob/main/EXAMPLES.md#exploiting-http-request-smuggling-to-reveal-front-end-request-rewriting)  

## Why ?

The  main objective is to help in the making of smuggle request

The goal is not to replace Burp, but to offer an alternative with CLi. This has the advantage to "automate" attack and so on

### *"HTTP Request Smugglin"* Kezako?

HTTP request smuggling is a technique for interfering with the way a web site processes sequences of HTTP requests. It was discover in 2005, and repopularized by PortSwigger's research.

It happends when users send requests to a front-end server (load balancer or reverse proxy) and this server forwards requests to one or more back-end servers.

When the front-end server forwards HTTP requests to a back-end server, it typically sends several requests over the same back-end network connection (efficient and performant). The protocol is very simple: HTTP requests are sent one after another, and the receiving server parses the HTTP request headers to determine where one request ends and the next one begins. **HTTP request smugging consist of luring backend server in its HTTP request parsing to make requests getting interpreted differently by the front-end and back-end systems** (failed to adequatly determine begins & ends of requets)


We have 3 possibilities:
* **CL.TE**: Front end uses `Content-Length` header and the back end uses `Transfer-Encoding`
* **TE.CL**: Front end: `Transfer-Encoding`, back end: `Content-Length`. (Fake `Content-Length`)
* **TE.TE**: Both server use `Transfer-Encoding` but one of those can be induced to not process it by obfuscating the header in some way

### Analyze `TE.CL` request treatment

As we want to see how a request is treated and thus how we could interfere this treatment, we will simulate it.

**First**, The front-end server treats the request using `Content-Length` header:
```shell
cat request | httpcustomhouse -cl
## Output is the request transmitted to front-end
```

**Then**, The front use `chunk` encoding to parse HTTP request:
```shell
cat request | httpcustomhouse -cl | httpcustomhouse -te
## Output is the request treated by back end
```

**Finally**, if we want to see if there is some part of `request` that hasn't been treated by backend and thus will be interpreted as the beginning of the next request:
```shell
cat request | httpcustomhouse -cl | httpcustomhouse -te -r
## Output is the request treated by back end
## Color output: the part of the request not treated by backend
## TRICK: add 2>&1 >/dev/null at the end to only obtain the non-treated part
```

### Analyze `CL.TE` request treatment

Same as [`TE.CL`](#analyze-tecl-request-treatment) idea:

First parse request using `chunk` encoding, then using `Content-Length` and finally print part of the request not already treated by back end:
```shell
cat request | httpcustomhouse -te | httpcustomhouse -cl -r
## Output is the request treated by back end
## Color output: the part of the request not treated by backend
## TRICK: add 2>&1 >/dev/null at the end to only obtain the non-treated part
```

## Building HTTP request

As `httpcustomhouse` use raw HTTP request as input you need to be able to construct it. There are several ways:
* Intercept request with `burp`, `mitmproxy` and save it to a file
* Use curl and an HTTP [`echo-server`](https://github.com/ariary/httpecho) to see sending request and save it to a file
* Take inspiration from the templates present in `samples` directory

**⚠️**: It is important to embed `\r` character and other special characaters in your request file. Edit request with an editor could withdraw them. use `cat -A` to see them. For example, in chunk encoding the final `0` must be followed by `\r\n\r\n`. 

### Use echo server

**First**, set up an echo server:
* Use **[mine](https://github.com/ariary/httpecho)** (compatible with request smuggling research)
* with `socat`
* with `netcat`
* Build your own

And then Make your `curl` request specifying your echo server as a proxy (the request won't reach the end server):
```shell
curl --proxy http://localhost:[port] ...REQUEST...
```

#### `Socat`

Constantly server + see `\r` character

The one-liner:
```shell
socat -v -v TCP-LISTEN:8888,crlf,reuseaddr,fork SYSTEM:"echo HTTP/1.0 200; echo Content-Type\: text/plain; echo; cat"
```


#### `netcat`

Serve 1 request + save it in a file

The one-liner:
```shell
nc -lp 8888 -c "tee myfile"
## or nc -nlvp 8888 > myfile  2>/dev/null &
```

#### Example

~ To build `samples/diff-te.cl`:
```shell
## Construct embedded request
nc -lp 8888 -c "tee data" #use another shell to make curl request
curl -X GET --proxy http://localhost:8888/  vulnerable-website.com -H 'Content-Length: 144' -H 'Content-Type: application/x-www-form-urlencoded'

## Construct the whole request
nc -lp 8888 -c "tee diff-te.cl" #use another shell to make curl request
curl -X POST --proxy http://localhost:8888/  vulnerable-website.com --data-binary "@data" -H 'transfer-encoding:chunked' -H 'Content-Length: 4'
rm data

```

## Send raw HTTP request

`httpclient` is the equivalent of **`curl` for raw request**.

**Why?**
* `curl` & go http client rewrite http request
* `ncat` and `openssl s_client` aren't fulle satisfying

```shell
cat [raw_request] | httpclient https://[URL]:[PORT]
```

### Alternatives
When you request is good, send it:
```Shell
cat [raw_request] | openssl s_client -ign_eof -connect [target_url]:443
#or use ncat from nmap package
cat [raw_request]| ncat --ssl [target_url] 443
```

Or if the target does not use tls/ssl:

```Shell
cat [raw_request] | nc -q 5 [target_url] 80 # or -w 5
```

## Install
```shell
# From Release:
curl -lO -L https://github.com/ariary/HTTPCustomHouse/releases/latest/download/httpcustomhouse && chmod +x httpcustomhouse

# With go:
go install github.com/ariary/HTTPCustomHouse/cmd/httpcustomhouse@latest
```
