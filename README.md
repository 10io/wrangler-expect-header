# wrangler-expect-header

This repository is for showing a problem between:

* [Cloudflare wrangler](https://github.com/cloudflare/workers-sdk).
* `PUT` requests with the [`Expect` header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Expect).

## Context

We have been using Cloudflare workers as a proxy in front of the main backend.

That main backend can receive uploads with the [`Expect` header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Expect) set.
With that header, the backend will reply a `100 Continue` response code in order to instruct the client the send the request body.

The main problem is that we get a [`workerd`](https://github.com/cloudflare/workerd) error:

> ✘ [ERROR] Uncaught (async) TypeError: Can't read from request stream after responding with an exception.

The assumption here is that sending the `100 Continue` response code is "interpreted" by workerd as "the response has been sent back", thus the request body can't be accessed.

## Reproduction steps

Apologies for the code quality level here, I'm not knowledgeable in Worker typescript code. I merely used the [hello world sample](https://developers.cloudflare.com/workers/get-started/guide/) from here that I modified to proxy requests.

This repository contains:

* A dummy `go` backend application that will simply dump the request body of the `PUT` request.
* A simple worker implementation that will simply proxy the request to the `go` backend.

Requirements:

* `go` installed and ready.
* the npm dependencies installed and ready (`$ npm install`).
* `curl` installed and ready.

Here are the steps.

1. Start the `go` backend server:
   ```shell
   $ cd backend
   $ go run main.go
   ```
   * This is now listening on port `8080`
1. In a different terminal, start the worker:
   ```shell
   $ npm wrangler dev
   ```


This is now ready. Let's try a few `PUT` requests.

Let's query the `go` backend server directly first. This is to assert the expected behavior.

```shell
$ curl -vvv --upload-file "./dummy.txt" "http://0.0.0.0:8080"
*   Trying 0.0.0.0:8080...
* Connected to 0.0.0.0 (0.0.0.0) port 8080
> PUT /dummy.txt HTTP/1.1
> Host: 0.0.0.0:8080
> User-Agent: curl/8.7.1
> Accept: */*
> Content-Length: 13
>
* upload completely sent off: 13 bytes
< HTTP/1.1 200 OK
< Date: Thu, 07 Nov 2024 08:06:50 GMT
< Content-Length: 37
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host 0.0.0.0 left intact
Hello, World! Received: Dummy content
```

Nothing special here. Let's add the `Expect` header:

```shell
$ curl -vvv -H "Expect: 100-continue" --upload-file "./dummy.txt" "http://0.0.0.0:8080"
*   Trying 0.0.0.0:8080...
* Connected to 0.0.0.0 (0.0.0.0) port 8080
> PUT /dummy.txt HTTP/1.1
> Host: 0.0.0.0:8080
> User-Agent: curl/8.7.1
> Accept: */*
> Expect: 100-continue
> Content-Length: 13
>
< HTTP/1.1 100 Continue
<
* upload completely sent off: 13 bytes
< HTTP/1.1 200 OK
< Date: Thu, 07 Nov 2024 08:07:38 GMT
< Content-Length: 37
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host 0.0.0.0 left intact
Hello, World! Received: Dummy content
```

As we can see here, `curl` is waiting for the `100 Continue` status code. `curl` then upload the file and the `go` backend "process" the request body.

Ok, now, let's try the same queries but _through_ the worker:

```shell
$ curl -vvv --upload-file "./dummy.txt" "http://0.0.0.0:8000"
*   Trying 0.0.0.0:8000...
* Connected to 0.0.0.0 (0.0.0.0) port 8000
> PUT /dummy.txt HTTP/1.1
> Host: 0.0.0.0:8000
> User-Agent: curl/8.7.1
> Accept: */*
> Content-Length: 13
>
* upload completely sent off: 13 bytes
< HTTP/1.1 200 OK
< Content-Length: 37
< Date: Thu, 07 Nov 2024 08:09:34 GMT
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host 0.0.0.0 left intact
Hello, World! Received: Dummy content
```

It's all fine. The worker logs show:

```
[wrangler:inf] PUT /dummy.txt 200 OK (4ms)
```

Ok, now, with the `Expect` header:

```shell
$ curl -vvv -H "Expect: 100-continue" --upload-file "./dummy.txt" "http://0.0.0.0:8000"
*   Trying 0.0.0.0:8000...
* Connected to 0.0.0.0 (0.0.0.0) port 8000
> PUT /dummy.txt HTTP/1.1
> Host: 0.0.0.0:8000
> User-Agent: curl/8.7.1
> Accept: */*
> Expect: 100-continue
> Content-Length: 13
>
< HTTP/1.1 503 Service Unavailable
< Content-Length: 125
< Content-Type: text/plain;charset=UTF-8
< Retry-After: 0
<
* HTTP error before end of send, stop sending
* abort upload
* Closing connection
Your worker restarted mid-request. Please try sending the request again. Only GET or HEAD requests are retried automatically.
```

Worker logs:

```
[wrangler:inf] PUT /dummy.txt 503 Service Unavailable (2ms)
✘ [ERROR] Uncaught (async) TypeError: Can't read from request stream after responding with an exception.
```

So the main question here: is it expected that Cloudflare Wrangler doesn't support `PUT` requests with the `Expect` header when the worker proxies the request.