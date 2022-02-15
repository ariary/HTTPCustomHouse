# HTTPCustomHouse

<div align=center>
<img src= https://github.com/ariary/HTTPCustomHouse/blob/main/img/E0D8F573-7824-42C1-BF6B-F58E5F14DB0E.png width=180>

<br><strong><i>CLi tools helping to forge  HTTP smuggling attack and others </i></strong>


<b>(<code>httpcustomhouse</code>)</b><br>
Analyze smuggled request without interacting with remote server. <sup><i><a href=#%EF%B8%8F-httpcustomhouse>(use it)</a></i></sup>

<b>(<code>httpoverride</code>)</b><br>
Manipulate HTTP raw request to sharpen attack. <sup><i><a href=#-httpoverride>(use it)</a></i></sup>

<b>(<code>httpclient</code>)</b><br>
Send HTTP raw request to perform the attack . <sup><i><a href=#-httpclient>(use it)</a></i></sup>

👁️ <strong>•</strong> 🔨 <strong>•</strong> 📬
</div> 

HTTP request smuggling is a technique for interfering with the way a web site processes sequences of HTTP requests ([more]()). The aim is to **perform request smuggling from command line**. It can't totally replace Burp Suite (or other GUI) but it proposes another approach, with more CLi. In order to offer a fully CLi experience while manipulating HTTP packets, these tools can be used with **[`httpecho`](https://github.com/ariary/httpecho)** which could help construct HTTP raw request. 

**Why That?**

* To learn
* Be able to solve challenge from CLi helps us to script resolution, automate exploit etc ...
* `curl`, go http client, `ncat`, `openssl s_client` aren't fully satisfying especially when dealing with "malformed http request"

**Real example:**
* [Forge `TE.CL` request smuggling attack](https://github.com/ariary/HTTPCustomHouse/blob/main/EXAMPLES.md#analyze-tecl-request-treatment)
* [Forge `CL.TE` request smuggling attack](https://github.com/ariary/HTTPCustomHouse/blob/main/EXAMPLES.md#analyze-clte-request-treatment)
* [Exploit `CL.TE`](https://github.com/ariary/HTTPCustomHouse/blob/main/EXAMPLES.md#exploiting-http-request-smuggling-to-reveal-front-end-request-rewriting)


## Usage

### 👁️ `httpcustomhouse`

*> allow you to reproduce HTTP request processing without interacting with online server*

**Show corresponding request treated by a server based on `Content-Length` Header treatment**:  
```shell
cat samples/te.cl | httpcustomhouse -cl
```
If the `Content-Length` is larger than the body size, the number of remaining bytes will be echoed

**Show corresponding request treated by a server based on chunk encoding treatment**:
```shell
cat samples/cl.te | httpcustomhouse -te
```

**Show the residue of the request that has not been treated** (in stderr):
```shell
cat samples/cl.te | httpcustomhouse -te -r
# -r (or --residue) works also for -cl
```

Demo: [ (🖼️) Visualize `TE.CL` ](https://github.com/ariary/HTTPCustomHouse/blob/main/img/hch.png)

### 🔨 `httpoverride`

*> help to modify http request*

**Override/Modify Header of an HTTP request**:
```shell
cat [raw_request] | httpoverride -H "Content-Length:55" -A "Host: spoofed.com"
# -A add header, -H override header

```
**Remove Header of an HTTP request**:
```shell
cat [raw_request] | httpoverride -H "Accept:" # or -H "Accept"
```


### 📬 `httpclient`
*> transmit HTTP request to server (HTTP client)*

**Send a HTTP raw request**:
```shell
cat [raw_request] | httpclient [protocol]:[url]:[port]  # port is falcultative https -> 443, http -> 80
```

**Send request and see response in browser**:
```shell
cat [raw_request] | httpclient -B [protocol]:[url]:[port]  # -Bc use cookie for future requests in browser
# Open browser and visit the link displayed
```

## Install
```shell
# From Release:
curl -lO -L https://github.com/ariary/HTTPCustomHouse/releases/latest/download/httpcustomhouse && chmod +x httpcustomhouse
curl -lO -L https://github.com/ariary/HTTPCustomHouse/releases/latest/download/httpoverride && chmod +x httpoverride
curl -lO -L https://github.com/ariary/HTTPCustomHouse/releases/latest/download/httpclient && chmod +x httpclient
# With go:
go install github.com/ariary/HTTPCustomHouse/cmd/httpcustomhouse@latest
go install github.com/ariary/HTTPCustomHouse/cmd/httpclient@latest
go install github.com/ariary/HTTPCustomHouse/cmd/httpoverride@latest
```


## *"HTTP Request Smuggling"* Kezako?

HTTP request smuggling is a technique for interfering with the way a web site processes sequences of HTTP requests. It was discover in 2005, and repopularized by PortSwigger's research.

It happends when users send requests to a front-end server (load balancer or reverse proxy) and this server forwards requests to one or more back-end servers.

When the front-end server forwards HTTP requests to a back-end server, it typically sends several requests over the same back-end network connection (efficient and performant). The protocol is very simple: HTTP requests are sent one after another, and the receiving server parses the HTTP request headers to determine where one request ends and the next one begins. **HTTP request smugging consist of luring backend server in its HTTP request parsing to make requests getting interpreted differently by the front-end and back-end systems** (failed to adequatly determine begins & ends of requets)


We have 3 possibilities:
* **CL.TE**: Front end uses `Content-Length` header and the back end uses `Transfer-Encoding`
* **TE.CL**: Front end: `Transfer-Encoding`, back end: `Content-Length`. (Fake `Content-Length`)
* **TE.TE**: Both server use `Transfer-Encoding` but one of those can be induced to not process it by obfuscating the header in some way


## Building HTTP request

As `httpcustomhouse` use raw HTTP request as input you need to be able to construct it. There are several ways:
* Intercept request with `burp`, `mitmproxy` and save it to a file
* Use curl and an HTTP [`echo-server`](https://github.com/ariary/httpecho) to see sent request and save it to a file ***(SUGGESTED)***
* Take inspiration from the templates present in `samples` directory

**⚠️**: It is important to embed `\r` character and other special characaters in your request file. Edit request with an editor could withdraw them. use `cat -A` to see them. For example, in chunk encoding the final `0` must be followed by `\r\n\r\n`. 

### Use echo server

**First**, set up an [echo server](https://github.com/ariary/httpecho):
```shell
httpecho -d raw
# will save request in "raw" file
``` 

And then Make your `curl` request specifying your echo server as a proxy (the request won't reach the end server):
```shell
curl --proxy http://localhost:[port] ...REQUEST...
```

#### Alternatives
##### `Socat`

Constantly server + see `\r` character

The one-liner:
```shell
socat -v -v TCP-LISTEN:8888,crlf,reuseaddr,fork SYSTEM:"echo HTTP/1.0 200; echo Content-Type\: text/plain; echo; cat"
```

##### `netcat`

Serve 1 request + save it in a file

The one-liner:
```shell
nc -lp 8888 -c "tee myfile"
## or nc -nlvp 8888 > myfile  2>/dev/null &
```


## Send raw HTTP request

As we deal with HTTP raw request we must be able to send them. `httpclient` is the equivalent of **`curl` for raw request**.

**Why?**
* `curl` & go http client rewrite http request (this is not satisfying for web pentest in general)
* `ncat` and `openssl s_client` aren't fully satisfying also

```shell
cat [raw_request] | httpclient https://[URL]:[PORT]
```

### Alternatives
When you request is good, send it:
```Shell
cat [raw_request] | openssl s_client -ign_eof -connect [target_url]:443
#or use ncat from nmap package
cat [raw_request]| ncat --ssl [target_url] 443
# or if target does not use tls/ssl
cat [raw_request] | nc -q 5 [target_url] 80 # or -w 5
```
