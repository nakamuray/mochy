# mochy - Mock HTTP Server

`mochy` is a mock HTTP server.

It do these things:

* return HTTP response specified by command-line argument (lua script)
* dump HTTP request it receive to terminal

```
Usage: mochy [options] <lua script>
  -addr string
        address to serve (default ":8080")
  -f string
        script filename
```

lua script is expeted to return:
* string, which is used as a response body
* table with key "code", "contentType" and "body"
* function, which return those things. this function is called at each request

## examples
return fixed string:
```
$ mochy 'return "hello world\n"'
```

return 404 response:
```
$ mochy 'return {code=404, body="sorry.\n"}'
```

return random numbers:
```
$ mochy 'return function() return math.random() .. "\n" end'
```

return back requrest path name:
```
$ mochy 'return function(req) return req.url.path .. "\n" end'
```
