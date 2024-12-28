# crazys-home-server

Home server for file storage and private web hosting

## Next-time

Implement JWT logic, without worrying about https/tls (will be configured later
or at the proxy level)

## Structure

It is important to note that the project should follow a client-server structure
instead of a monolithic one. This is done to separate the concerns for the server
and the client.

By adding authentication and authorization to the application, its complexity grows.
The client will serve as the front-end. It will:
- Provide users with a UI to fill in forms
- Store authentication keys provided by the server
- Make calls to the server

The server, on the other hand, will be a RESTful API server. It will:
- Provide an endpoint for logging in (POST)
    - Generating tokens, etc.
- Provide endpoints for uploading files (POST)
- Provide endpoints for downloading files (GET)

The client will still technically be a *server*, however it will simply route the
user's requests to the back-end (the RESTful API server)

## Security

The internal servers (i.e. the file storage server) will not be exposed to the
outside world. They will serve on ports that are not port-forwarded. Therefore,
they will operate in HTTP, NOT HTTPS. This will improve performance and simplify
their implementation.

Security (HTTPS/TLS) will be handled by a load-balancer/reverse proxy, which 
**will be exposed to the outside world**. For learning purposes, a simple PoC
will be written in Go, however, for production, a more well-established
solution, such as NGINX, will be used (*after understanding what it does and why
it is used*)



## References

- [Go/Gin getting started docs](https://go.dev/doc/tutorial/web-service-gin)
- [Gin upload docs](https://gin-gonic.com/docs/examples/upload-file/multiple-file/)
- [Gin static files](https://chroniconl.vercel.app/articles/serving-static-content-with-go-and-gin)

## Resources to look up later

- [Go REST client, with appended Authorization header](https://dev.to/der_gopher/writing-rest-api-client-in-go-3fkg#:~:text=req.Header.Set(%22Authorization%22%2C%20fmt.Sprintf(%22Bearer%20%25s%22%2C%20c.apiKey)))
- [Go Authentication methods by JetBrains](https://www.jetbrains.com/guide/go/tutorials/authentication-for-go-apps/auth/)
- [Go HTTPS/TLS](https://eli.thegreenplace.net/2021/go-https-servers-with-tls/)
- [Go net/http/httputil reverse proxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)
