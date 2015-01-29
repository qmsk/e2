import asyncio
import logging; log = logging.getLogger('qmsk.net.udp')
import socket

import qmsk.net.socket

class Error (Exception):
    pass

class DatagramSocket (qmsk.net.socket.Socket):
    """
        UDP socket
    """

    def __init__ (self, *args, **opts):
        super().__init__(*args, nonblocking=True, **opts)

    def send (self, buf):
        """
            Send packet with buf to our connected remote peer.

            This will fail if this is not a connect()'d socket.
        """

        log.debug("%s: %d...", self, len(buf))
        
        # XXX: SOCK_DGRAM never raises BlockingIOError?
        # TODO: just drop packets if no send buffer
        return self.sock.send(buf)

    def sendto (self, addr, buf):
        """
            Send packet with buf to given remote peer.

            Intended for use with listen()'d sockets.
        """
        log.debug("%s: %s: %d...", self, qmsk.net.socket.addrname(addr), len(buf))
        
        # XXX: SOCK_DGRAM never raises BlockingIOError?
        # TODO: just drop packets if no send buffer
        return self.sock.sendto(buf, addr)

    @asyncio.coroutine
    def recv (self, size):
        """
            Receive data from addr=(host,port).

            Returns bytes.

            Raises socket.error on network errors.
            Raises Error on semantical errors, such as truncated packets.
        """

        # detect truncated packets, when the packet on the wire was larger than our recv size
        # XXX: use something that supports recv() flags=MSG_TRUNC, and returns addr
        buf = yield from self.loop.sock_recv(self.sock, size)
        
        # XXX: requires MSG_TRUNC
        if len(buf) > size:
            raise Error("{self}: truncated message at {read}/{len} bytes".format(self=self, read=read_size, len=len(buf)))

        return buf
        
    @asyncio.coroutine
    def end (self):
        """
            Closes the sockets
        """ 
        self.sock.close()
        

@asyncio.coroutine
def connect (host, port, loop=None, listen_host=None, listen_port=None, **opts):
    """
        Open a new UDP socket bound to some arbitrary local port, and connected to the given host/port.
        
        Raises Error on name resolution or socket errors.
    """

    if loop is None:
        loop = asyncio.get_event_loop()
    
    # remote addr
    try:
        connect_ai = yield from loop.getaddrinfo(host, port, type=socket.SOCK_DGRAM)
    except socket.gaierror as error:
        raise Error("%s:%s: %s" % (host, port, error))
    
    # local addr
    if listen_host or listen_port:
        try:
            bind_ai = yield from loop.getaddrinfo(listen_host, listen_port, type=socket.SOCK_DGRAM, flags=socket.AI_PASSIVE)
        except socket.gaierror as error:
            raise Error("%s:%s: %s" % (local_host, local_port, error))
    else:
        bind_ai = [ ]

    # TODO: except erorrs and try further alternatives, not just the first result?
    for family, type, proto, connect_canonname, connect_sockaddr in connect_ai:
        log.debug("socket family=%s type=%s: connect sockadr=%s", family, type, connect_sockaddr)
        
        try:
            sock = socket.socket(family, type, proto)
        except socket.error as error:
            raise Error("%s:%s: socket(family=%d, type=%d, proto=%d): %s" % (host, port, family, type, proto, errror))

        for bind_family, bind_type, bind_proto, bind_canonname, bind_sockaddr in bind_ai:
            log.debug("socket family=%s type=%s: bind sockaddr=%s", bind_family, bind_type, bind_sockaddr)

            if bind_family == family and bind_type == type and bind_proto == proto:
                # TODO: except and try next
                try:
                    sock.bind(bind_sockaddr)
                except socket.error as error:
                    raise Error("%s:%s: bind(%s): %s" % (host, port, bind_sockaddr, error))

        # we assume this is non-blocking for UDP sockets..
        try:
            sock.connect(connect_sockaddr)
        except socket.error as error:
            raise Error("%s:%s: connect(%s): %s" % (host, port, connect_sockaddr, error))

        return DatagramSocket(sock, loop=loop, **opts)

@asyncio.coroutine
def listen (port, host=None, loop=None, **opts):
    """
        Open a new UDP socket bound to the given local port, and not connected to any remote.

        Raises Error on name resolution or socket errors.
    """

    if loop is None:
        loop = asyncio.get_event_loop()

    try:
        ai = yield from loop.getaddrinfo(host, port, type=socket.SOCK_DGRAM, flags=socket.AI_PASSIVE)
    except socket.gaierror as error:
        raise Error("%s:%s: %s" % (host, port, error))

    # TODO: except erorrs and try further alternatives, not just the first result?
    for family, type, proto, canonname, sockaddr in ai:
        log.debug("socket family=%s type=%s: bind sockaddr=%s", family, type, sockaddr)

        try:
            sock = socket.socket(family, type, proto)
        except socket.error as error:
            raise Error("%s:%s: socket(family=%d, type=%d, proto=%d): %s" % (host, port, family, type, proto, error))

        try:
            sock.bind(sockaddr)
        except socket.error as error:
            raise Error("%s:%s: bind(%s): %s" % (host, port, bind_sockaddr, error))

        return DatagramSocket(sock, loop=loop, **opts)

