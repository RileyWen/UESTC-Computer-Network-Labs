package main

import (
    "fmt"
    "net"
    "net/http"
    "log"
    "bufio"
    "strings"
    "io"
    //"io/ioutil"
    //"os"
)

func hostname_to_ip_port(hostname string, index int) (ip, port string) {

    // If hostname contains port part, e.g. "baidu.com:80", do nothing.
    // Otherwise, append the default port "80" to hostname
    hostname_port_slice := strings.Split(hostname, ":")
    if len(hostname_port_slice) == 1 {
        hostname_port_slice = append(hostname_port_slice, "80")
    }
    //fmt.Printf("hostname_port_slice: %+v\n\n", hostname_port_slice)

    // Resolve the hostname to IP
    addrs_in_bytes, err := net.LookupIP(hostname_port_slice[0])
    checkError(err)
    ip = addrs_in_bytes[0].String()
    fmt.Printf("[#%d] IP resolverd from hostname: %s\n\n", index, ip)

    // Return IP and port
    return ip, hostname_port_slice[1]

}

func proxy_thread(conn net.Conn, index int) {
    // Close the socket when stepping out `proxy_thread` function
    defer conn.Close()

    conn_reader := bufio.NewReader(conn)
    conn_writer := bufio.NewWriter(conn)

    // Read HTTP Header from client
    request, err := http.ReadRequest(conn_reader)

    // If client close unexpectedly and EOF is met, just skip this HTTP request
    if err == io.EOF {
        fmt.Printf("[#%d] EOF Received.... End this routine...\n\n", index)
        return
    }

    fmt.Printf("[#%d] Request Header:\n%+v\n\n",    index, *request)
    fmt.Printf("[#%d] Request Schenme: %s\n\n",       index, request.URL.Scheme)
    fmt.Printf("[#%d] Request Host: %s\n\n",          index, request.URL.Host)
    fmt.Printf("[#%d] Request Path: %v\n\n",          index, request.URL.Path)

    //if request.URL.Scheme != "http" {
    //    fmt.Println("Only support 'http' protocol!")
    //    return
    //}

    // Resolve the hostname that client requests to IP
    ip, port := hostname_to_ip_port(request.URL.Host, index)

    // Open TCP socket to the host
    forward_conn, err := net.Dial("tcp", ip + ":" + port)
    checkError(err)

    // Close this socket when stepping out `proxy_thread` function
    defer forward_conn.Close()

    forward_conn_reader := bufio.NewReader(forward_conn)
    forward_conn_writer := bufio.NewWriter(forward_conn)

    // Modify the HTTP header (mainly in URL part) in order to request the host
    forward_request := request
    forward_request.URL.Host = ""
    forward_request.URL.Scheme = ""

    // Send HTTP request header to host
    forward_request.Write(forward_conn_writer)
    forward_conn_writer.Flush()

    // Read HTTP response from host
    forward_response, err := http.ReadResponse(forward_conn_reader, nil)
    checkError(err)
    fmt.Printf("[#%d] Forward Response: %+v\n\n", index, *forward_response)

    fmt.Printf("[#%d] Sending Response back to Client...\n\n", index)
    // Forward the HTTP response back to client
    forward_response.Write(conn_writer)
    conn_writer.Flush()


    fmt.Printf("[#%d] End of Handling\n\n", index)

    // Caching files of real server, NOT IMPLEMENTED
    //file_content, err := ioutil.ReadFile("files"+request.URL.Path)

    //switch {
    //case os.IsNotExist(err):
    //    fmt.Printf("File Not Exists! Request %s for it...\n\n", forward_addr)

    //

    //default:
    //    fmt.Printf("File content:\n%s\n\n", file_content)

    //}



}

func main() {
    // The addr and port on which our proxy will be listening on
    serv_addr := "localhost:10086"

    // Open TCP socket
    listener, err := net.Listen("tcp", serv_addr)
    checkError(err)

    // Close the socket when main function finishes
    defer listener.Close()

    fmt.Printf("Server is listening at %v\n\n", serv_addr)

    i := 0

    // Dead Loop for accepting incoming TCP connections
    for {
        // Accept new TCP connection
        conn, _ := listener.Accept()
        i++

        fmt.Printf("Accepted connection from %v\n\n", conn.RemoteAddr())
        fmt.Printf("And the HTTP Request is sent to %v\n\n", conn.LocalAddr())

        // Dispatch the client HTTP request to go-routines (i.e. user-level thread)
        go proxy_thread(conn, i)
    }

}

func checkError(err error) {
    if err != nil {
        // If some errors that are not expected occurs, just let it crash
        log.Fatal(err)
    }
}
