# Usage
1. Go into directory `/cmd/httpserver` and run the command `go run .`
 
    This will spawn an instance of the HTTP server running on port 8080 (this port number can be easily changed in `httpserver/main.go`).

2. Review the different server handlers under `/internal/handlers/handlers.go`. The default server handler will show a video when you visit <http://localhost:8080/video>

    This showcases how the HTTP protocol can handle the transportation of a multitude of different data types, which in this case is a stream of binary bits.

3. Change which server handler is used by commenting out the default server handler and 
removing the comments for one of the following provided server handlers:

    `srv, err := server.Serve(handlers.Handler, port)`

    `srv, err := server.Serve(handlers.ProxyHandlerWithTrailers, port)`

    `srv, err := server.Serve(handlers.ProxyHandler, port)`

4. Explore the different behaviors gained from changing which server handler is used:

    - Use of the regular `handlers.Handler` will return different responses based on which path is visited. Available paths include the root path (`/`) and two problem paths (`/yourproblem` & `/myproblem`). Viewing of these responses can be done through the browser or a command line utility like `curl`.

    - The two other server handlers are very similar (`handlers.ProxyHandlerWithTrailers` & `handlers.ProxyHandler`). View these handlers' proxy behavior by visiting <http://localhost:42069/httpbin/stream/100> with either handler active. This will stream back 100 lines of JSON-encoded request data. Can also view it with the `curl --raw http://localhost:42069/httpbin/stream/100` command. 
        - Both handlers make a request to the [HTTPBin](https://httpbin.org/) service and return a list of JSON data, thereby acting as a proxy (you make a request to the server handler, which then makes a request to HTTPBin, which then returns to you the response from HTTPBin). They also use [chunked encoding](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Transfer-Encoding) when sending this response back, which means that instead of waiting for the entire response to come back from HTTPBin, they will instead send chunks of data to the client until the response is finished.
        - The only real different between the two handlers is the use of [HTTP trailers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Trailer). This feature in the HTTP protocol is useful for sending clients information about the data being sent (known as metadata) which can not be known until all the chunks of data that make up the response is gathered. A great example of this is a [checksum](https://algomaster.io/learn/system-design/checksums).
