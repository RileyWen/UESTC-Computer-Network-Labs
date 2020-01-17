from socket import *


def main():
    # Create TCP socket
    with socket(AF_INET, SOCK_STREAM) as listener:
        # bind ‘127.0.0.1：1230’
        listener.bind(("localhost", 1230))
        # Start Listening
        listener.listen(1)

        while True:
            try:
                # Ready to accept connection and generate connection socket
                conn, addr = listener.accept()
                with conn:
                    print("Connected from: ", addr)

                    # Recv HTTP Header
                    header = conn.recv(8092)

                    print("Raw Header:")
                    print(header)

                    # If Header is empty, which indicates ERROR, just skip it
                    if not header:
                        continue

                    # Split header line by line
                    header_lines = header.decode().splitlines()

                    try:
                        print("Gonna open: %s" % header_lines[0].split()[1][1:])

                        # Open the file according to URL in HEADER
                        file = open(header_lines[0].split()[1][1:])
                        # Read file
                        raw_data = file.read()
                    except FileNotFoundError:
                        # If the file doesn't exist, return 404 Not Found
                        conn.send(b"HTTP/1.1 404 Not Found\r\n\r\n")
                    else:
                        # If file exists，return 200 OK and file content
                        conn.send(b"HTTP/1.1 200 OK\r\n\r\n")
                        conn.send(raw_data.encode())
                        file.close()
            except InterruptedError:
                print("Gracefully Exited.")


if __name__ == '__main__':
    main()
