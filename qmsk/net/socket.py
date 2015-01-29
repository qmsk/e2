import errno
import logging; log = logging.getLogger('qmsk.net.socket')
import socket

def normhost (host):
    """
        Normalize an inet host address, handles IPv6, IPv4 and IPv4-in-IPv6 addresses.
    """
        
    if host.startswith('::ffff:'):
        host = host.split('::ffff:', 1)[1]

    return host

class Socket (object):
    """
        A single socket.
    """

    def __init__ (self, sock, loop=None, nonblocking=None):
        """
            loop:       asyncio.BaseEventLoop
            sock:       socket.socket
        """

        if loop is None:
            loop = asyncio.get_event_loop()

        self.sock = sock
        self.loop = loop

        if nonblocking is not None:
            self.sock.setblocking(not nonblocking)

    def fileno (self):
        return self.sock.fileno()

    def sockname (self):
        """
            Socket local address as a human-readble string.
        """

        return addrname(self.sock.getsockname())
    
    def sockhost (self):
        """
            Socket local host address as a human-readble string.
        """
        
        sockaddr = self.sock.getsockname()

        return normhost(sockaddr[0])
    
    def sockport (self):
        """
            Return int: local socket bind() port.
        """

        sockaddr = self.sock.getsockname()

        # XXX: not all sockets have a port... e.g. AF_UNIX
        return sockaddr[1]

    def peername (self):
        """
            Socket remote address as a human-readble string.
        """

        return addrname(self.sock.getpeername())
    
    def peerhost (self):
        """
            Return the IPv4/IPv6 level address for the remote peer.

            Returns a string.
        """

        sockaddr = self.sock.getpeername()

        return normhost(sockaddr[0])

    def peerport (self):
        """
            Return the transport (TCP/UDP) level port for the remote peer.

            Returns an integer.
        """

        sockaddr = self.sock.getpeername()

        # XXX: not all sockets have a port... e.g. AF_UNIX
        return sockaddr[1]

def addrname (addr):
    """
        Return a human-readable representation of the given socket address.
    """

    # will not block with numerics
    host, port = socket.getnameinfo(addr, socket.NI_DGRAM | socket.NI_NUMERICHOST | socket.NI_NUMERICSERV)

    return "{host}:{port}".format(host=normhost(host), port=port)
