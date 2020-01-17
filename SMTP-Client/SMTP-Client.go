package main

import (
    "bufio"
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "net"
    "os"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s your-gmail-address receiver-email-address", os.Args[0])
        os.Exit(1)
    }

    mail_server_addr := "smtp.qq.com:587"

    // Resolve the mail server address
    tcpAddr, err := net.ResolveTCPAddr("tcp", mail_server_addr)
    checkError(err)

    // Open TCP socket to QQ mail server
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    checkError(err)

    var resp []byte

    // Receive message from server first
    resp = recvFromServer(conn)
    checkResponse(resp[0:], "220")

    // Send EHLO command to check what extension server supports and print server response
    sendToServer(conn, "EHLO RileyWen\r\n")
    resp = recvFromServer(conn)
    checkResponse(resp[0:], "250")

    // Send STARTTLS command and print server response
    sendToServer(conn, "STARTTLS\r\n")
    resp = recvFromServer(conn)
    checkResponse(resp[0:], "220")

    // Wrap the connection with TSL for security
    tlsConn := tls.Client(conn, &tls.Config{
        ServerName: "smtp.qq.com",
    })

    // Send EHLO command again (through TLS) and print server response
    sendToServer(tlsConn, "EHLO RileyWen\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "250")

    // Send AUTH LOGIN command
    sendToServer(tlsConn, "AUTH LOGIN\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "334")

    // Send username encoded in base64
    username := base64.StdEncoding.EncodeToString([]byte(os.Args[1]))
    sendToServer(tlsConn, username+"\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "334")

    // Read password
    fmt.Print("Enter Password: ")
    reader := bufio.NewReader(os.Stdin)
    password, err := reader.ReadString('\n')
    checkError(err)

    // Send password encoded in base64
    password = base64.StdEncoding.EncodeToString([]byte(password))
    sendToServer(tlsConn, password+"\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "235")

    // Send MAIL FROM command and print server response
    fromEmail := os.Args[1]
    sendToServer(tlsConn, "MAIL FROM: <"+fromEmail+">\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "250")

    // Send RCPT TO command and print server response
    toEmail := os.Args[2]
    sendToServer(tlsConn, "RCPT TO: <"+toEmail+">\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "250")

    // Send DATA command and print server response
    sendToServer(tlsConn, "DATA\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "354")

    // Send message data
    text := "This is a test email from rileywen"
    message := "From: " + fromEmail + "\r\n"
    message += "To: " + toEmail + "\r\n"
    message += "Subject: " + text + "\r\n\r\n"
    message += text + "\r\n"
    sendToServer(tlsConn, message)

    // End message with a single period
    sendToServer(tlsConn, ".\r\n")
    resp = recvFromServer(tlsConn)
    checkResponse(resp[0:], "250")

    // Send QUIT command and print server response
    sendToServer(tlsConn, "QUIT\r\n")

    // close connection
    tlsConn.Close()

    os.Exit(0)
}

func sendToServer(conn net.Conn, data string) {
    // Write to socket
    _, err := conn.Write([]byte(data))
    fmt.Printf("[Client]\n%s\n", data)
    checkError(err)
}

func recvFromServer(conn net.Conn) []byte {
    var resp [512]byte

    // Read from socket
    n, err := conn.Read(resp[0:])
    checkError(err)
    fmt.Println("[Server]\n" + string(resp[0:n]))

    return resp[0:]
}

func checkResponse(resp []byte, code string) {
    // If the response code is not `code` that we expected, let it crash
    if string(resp[0:3]) != code {
        fmt.Fprintf(os.Stderr, "Code %s is EXPECTED but not received from server.", code)
        os.Exit(1)
    }
}

func checkError(err error) {
    // If the some unexpected errors occur, let it crash
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}
