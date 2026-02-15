---
title: SSH Tunneling — Just Layered Transport Engineering
date: 2026-02-15
image: 
tags: distributed-systems, networking, ssh, infra
summary: SSH tunneling is not magic routing or a lightweight VPN — it is encrypted, multiplexed socket forwarding implemented at the application layer. The SSH client binds to a local port, accepts normal TCP connections, and bridges those byte streams through encrypted channels to a remote host, where new TCP connections are opened on our behalf.
---

We use SSH tunneling as a convenience tool:

```shell
ssh -i ~/.ssh/ke-bastion.pem \
    -L 5432:db.internal:5432 \
    ec2-user@203.0.113.10
```

Suddenly, `localhost:5432` talks to a private database inside a VPC.

It feels like teleportation.

SSH tunneling is not magic. It is carefully **layered transport mechanism** built on top of a multiplexed, encrypted TCP session.
And understanding how it works clarifies a lot about networking, trust boundaries, and protocol design.

### SSH Is a Multiplexed Transport, Not Just a Shell
We casually think of SSH as "remote terminal access". That's the UI.
Underneath, SSH is a binary protocol running over TCP (usually port 22) that provides:
- Key exchange and authentication
- Symmetric encryption
- Integrity protection
- Optional compression
- **Multiplexed logical channels**

That last capability is what makes tunneling possible.

After the handshake completes, SSH establishes a secure session. 
Inside that single TCP connection, multiple logical channels can exist simultaneously — interactive shells, exec sessions, X11 forwarding, and importantly, TCP port forwarding.

SSH is closer to a secure stream multiplexer than a remote shell tool.

### What Local Port Forwarding Actually Does
Consider the common case:
```shell
ssh -L 5432:db.internal:5432 user@bastion
```

At a high level, this binds a local port and forwards traffic to a remote host through the SSH connection. But internally, the sequence matters.
1. **SSH Connection**
    - Local machine establishes a TCP connection to `bastion:22`
    - SSH Handshake completes
    - Encryption keys are established
2. **Local Listener**
    - The SSH client binds to `localhost:5432` (_It simply inserts itself as a TCP server locally._)
    - Now our OS thinks a service is listening on port 5432. (_The OS believes it’s delivering bytes to a local service.
      In reality, SSH forwards those bytes elsewhere._)
3. **Incoming Local Connection**
    - When our application connects to that local port, the SSH client:
       - Accepts the TCP connection.
       - Opens a new SSH channel inside the encrypted session.
       - Sends a request to the server to open a TCP connection to `db.internal:5432`
4. **Remote TCP Connect**
    - The SSH server (on bastion) then:
      - Opens its own TCP connection to `db.internal:5432`
      - Bridges that socket to the SSH channel

The data path becomes:
```shell
    Local App
       ↓
    Local TCP (localhost:5432)
       ↓
    SSH Client
       ↓ (encrypted SSH channel)
    Bastion SSH Server
       ↓
    Remote TCP (db.internal:5432)
       ↓
    Database
```

Every bite is framed into SSH channel packets, encrypted, transmitted over the outer TCP connection, decrypted, and written to the remote socket.

This is TCP inside TCP, carried over an encrypted session with channel-level multiplexing.

### The Cost of TCP-over-TCP
SSH tunneling encapsulates one TCP stream inside another.

The outer layer (client to bastion) is TCP. <br>
The inner layer (bastion to destination) is also TCP.

This layering introduces subtle trade-offs:
- Retransmissions happen at both layers.
- Packet loss on the outer connection cans stall all inner streams
- Head-of-line blocking affects multiplexed channels.

In stable networks, this overhead is negligible. In lossy or high-latency environments, TCP-over-TCP can degrade throughput significantly.

SSH tunneling optimises for security and simplicity, not for high-performance bulk transport.

### Flow Control and Isolation
Because SSH multiplexes many channels over one connection, it must implement per-channel flow control.
Each channel advertises a window size, preventing a single forward stream from consuming unbounded memory or starving others.

Without this mechanism, one busy database dump through a tunnel could block interactive shells or other forwarded connections sharing the same SSH session.

This channel-level backpressure is what allows SSH to safely multiplex heterogeneous traffic over one encrypted transport.

### Encryption Scope and Trust Boundaries
SSH encrypts traffic between the client and the SSH server. That is the boundary of confidentiality.

If we tunnel PostgreSQL through a bastion, the traffic is encrypted between local machine and the bastion. 
Beyond that hop, unless PostgreSQL itself uses TLS, the traffic is plaintext.

SSH tunneling does not automatically provide end-to-end encryption. It secures a transport segment.

Understanding this is critical in production environments where multiple trust zones exists.

SSH tunneling effectively punches holes through network boundaries. It allows:
- Access to private VPC resources
- Traversal across firewalls
- Bypassing of direct routing restrictions

From a security standpoint, this is powerful and dangerous.

### Remote and Dynamic Forwarding: Same Primitive, Different Direction
**Remote port forwarding (-R)** simple reverses the initiation direction. Instead of binding locally, the SSH server binds a port remotely.
Incoming connections there are tunneled back to the client.

This is particularly useful in NAT-restricted environments where only outbound connections are allowed.
A reverse tunnel effectively creates an inbound path by piggybacking on an outbound SSH session.

**Dynamic Forwarding (-D)** goes one step further. Instead of hardcoding a destination, SSH acts as a SOCKS proxy.
The client listens locally, interprets the SOCKS handshake to determine the requested destination, and opens a new SSH channel for each connection.

At that point, SSH becomes a generic encrypted TCP proxy. Any TCP-based application that can speak through SOCKS can route through tunnel.

---
### What SSH Tunneling Really Is
SSH tunneling is:

A secure, multiplexed transport layer that forwards arbitrary TCP connections through an encrypted session by mapping them onto logical channels.

It is not a VPN. <br>
It is not L3 routing. <br>
It is not protocol-aware proxying.

It is application-layer TCP encapsulation with encryption and channel multiplexing.
