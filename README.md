# crazys-home-server

Home server for file storage and private web hosting

## Next-time

Login

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

### Password hashing

Passwords are **hashed on the server-side**.

*If a hash is calculated on the client, the client authenticates to the server by
submitting their hash. The server then compares the hash to the database entry.
This means that if the database is exposed, attackers can authenticate as anyone 
by submitting the correct hash. Even though they cannot determine the original passwords,
they can still use the hashes directly to break authentication. With client-side hashing,
the hash effectively becomes the password.*
[reference](https://www.sjoerdlangkemper.nl/2020/02/12/the-case-for-client-side-hashing-logging-passwords-by-mistake/#:~:text=If%20a%20hash,becomes%20the%20password.)

It can be a good future exercise to implement a more robust way of sending over
sensitive information, such as hashing in the client and on the server, or adding
custom salt to the hash that only the server knows how to remove/parse out.

## Access

- Each user has access to their own files (in their own directory)
- Users can grant access to other users via the other person's username/id
    - People with view/download access can be added to each user's internal list
    of trusted users
- The user information will be checked on the back-end
- The front-end will purely serve as a means to display and handle the current
user's requests, and will not do any authentication logic besides retrieving and
saving JWT from the back-end

### Q: Why have passwords instead of just providing a JWT that is valid for N amount of time?

A: Users would then be liimited to a single device. They would need the JWT stored
somewhere, which is inconvenient. Instead, they can use credentials (username/password)

## References

- [Go/Gin getting started docs](https://go.dev/doc/tutorial/web-service-gin)
- [Gin upload docs](https://gin-gonic.com/docs/examples/upload-file/multiple-file/)
- [Gin static files](https://chroniconl.vercel.app/articles/serving-static-content-with-go-and-gin)
- [Go JWT Library create tokens](https://golang-jwt.github.io/jwt/usage/create/)
- [Authentication example](https://mazle78.notion.site/Authentication-and-Authorization-in-Go-with-Gin-120fd437022f80fbab95ff24bf9f0631)
- [PostgreSQL Go driver](https://github.com/jackc/pgx/tree/master)

## Resources to look up later

- [Go REST client, with appended Authorization header](https://dev.to/der_gopher/writing-rest-api-client-in-go-3fkg#:~:text=req.Header.Set(%22Authorization%22%2C%20fmt.Sprintf(%22Bearer%20%25s%22%2C%20c.apiKey)))
- [Go Authentication methods by JetBrains](https://www.jetbrains.com/guide/go/tutorials/authentication-for-go-apps/auth/)
- [Go HTTPS/TLS](https://eli.thegreenplace.net/2021/go-https-servers-with-tls/)
- [Go net/http/httputil reverse proxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- [Login system PoC](https://medium.com/@cheickzida/golang-implementing-jwt-token-authentication-bba9bfd84d60#:~:text=Implementing%20a%20Login%20System)

---

- [Auth w/ db example](https://ututuv.medium.com/building-user-authentication-and-authorisation-api-in-go-using-gin-and-gorm-93dfe38e0612)
- Using an .env file for things such as the JWT signing key
