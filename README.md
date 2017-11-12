# mochy - Mock HTTP Server

`mochy` is a mock HTTP server.

It just do these things:

* return HTTP response specified by command-line argument
* dump HTTP request it receive to terminal

```
Usage of mochy:
  -addr string
        address to serve (default ":8080")
  -body string
        response body (default "It's Works!\n")
  -code int
        status code (default 200)
  -content-type string
         (default "text/plain; charset=utf-8")
  -file string
        file path to use as a response body
```
