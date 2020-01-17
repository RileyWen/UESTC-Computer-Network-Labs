#!/usr/bin/env python3

import sys
import argparse

import time
import socket
from socket import socket as Socket


def main():
    # Create args parser
    parser = argparse.ArgumentParser()

    # add '-p' arg, which specify the port the server will be listening at
    # and the client will connect to
    parser.add_argument('--server-port', '-p', default=2081, type=int,
                        help='Server_Port to use')

    # add '-s' arg, which determine the program will run server or client
    parser.add_argument('--run-server', '-s', action='store_true',
                        help='Run a ping server')

    # add '--server-address' arg, which specify the addr that client will connect to
    parser.add_argument('--server-address', default='localhost',
                        help='Server to ping, no effect if running as a server.')

    args = parser.parse_args()

    # If '-s' is specified, run server, or run client
    if args.run_server:
        return run_server(args.server_port)
    else:
        return run_client(args.server_address, args.server_port, )


def run_server(server_port):
    # Create UDP socket
    with Socket(socket.AF_INET, socket.SOCK_DGRAM) as server_socket:

        # Set 'SO_REUSEADDR' so that we can reuse the same port immediately
        server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

        # Bind the port to listening socket
        server_socket.bind(('', server_port))

        # Starting receiving UDP packet
        print("Ping server ready on port", server_port)
        while True:
            # If receive the 'ping!'
            msg, client_address = server_socket.recvfrom(1024)
            if msg == b'ping!':
                server_socket.sendto(b"pong!", client_address)


def run_client(server_address, server_port):
    # Measure RTT for 10 times
    for i in range(0, 10):
        # Create UDP socket
        with Socket(socket.AF_INET, socket.SOCK_DGRAM) as client_socket:

            # set timeout to 1s
            client_socket.settimeout(1)

            # Specify the addr and port to connect to
            client_socket.connect((server_address, server_port))

            print('#%d ' % i, end='')

            # Measure RTT
            start = time.time_ns()

            # Send UDP ping packet
            client_socket.send(b'ping!')

            try:
                recv_ = client_socket.recv(4096)
            except socket.timeout:
                recv_ = None

            end = time.time_ns()
            elapsed = end - start

            # If pong packet is received normally
            if recv_ == b'pong!':
                print('RTT: %.2fms' % (elapsed / (10 ** 6)))
            else:
                print('Request Timeout')


if __name__ == "__main__":
    sys.exit(main())
