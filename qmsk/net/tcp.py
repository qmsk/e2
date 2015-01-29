"""
    asyncio TCP Server implementation.
"""

import asyncio
import logging; log = logging.getLogger('qmsk.net.tcp')
import socket

import qmsk.net.socket

LISTEN_BACKLOG = 16

class Error (Exception):
    pass

class SocketError (Error):
    pass

class StringError (Error):
    pass

class Stream (qmsk.net.socket.Socket):
    """
        Trivial asyncio-enabled TCP socket.
    """
    
    # split lines by given byte
    READLINE = b'\n'
    WRITELINE = b'\r\n'

    # strip lines by given bytes
    READLINES = b'\r\n'
    
    # optimum recv() length
    CHUNK = 4096

    # buf <-> string encoding
    ENCODING = 'utf-8'

    def __init__ (self, *args, encoding=None, readline=None, readlines=None, writeline=None, **opts):
        super().__init__(*args, nonblocking=True, **opts)
        
        # buffering for readline()
        # not used as a read() buffer
        self.buf = bytes()

        # string decoding for readline/writeline
        self.encoding = encoding or self.ENCODING
        self._readline = readline or self.READLINE
        self._readlines = readlines or self.READLINES
        self._writeline = writeline or self.WRITELINE

    @asyncio.coroutine
    def write (self, buf):
        """
            Write given data to socket.
        """

        log.debug("%d...", len(buf))
        
        try:
            return self.loop.sock_sendall(self.sock, buf)
        except socket.error as error:
            raise SocketError(error)

    @asyncio.coroutine
    def writestr (self, str):
        try:
            buf = str.encode(self.encoding)
        except UnicodeEncodeError as error:
            log.warning("%s: %r", error, line)
            raise StringError(error)
        
        log.debug("%s", buf)

        try:
            return self.loop.sock_sendall(self.sock, buf)
        except socket.error as error:
            raise SocketError(error)

    @asyncio.coroutine
    def writeline (self, line):
        """
            Write given line to socket.
        """
        
        try:
            buf = line.encode(self.encoding) + self._writeline
        except UnicodeEncodeError as error:
            log.warning("%s: %r", error, line)
            raise StringError(error)

        log.debug("%s", buf)

        try:
            return self.loop.sock_sendall(self.sock, buf)
        except socket.error as error:
            raise SocketError(error)

    @asyncio.coroutine
    def _recv (self, size_hint=0):
        """
            Reads any amount of data from the socket into our buf.

                size_hint:  hint on expected amount of data waiting to be read
            
            Returns True if any amount of data has been read.
            Returns False on EOF, no more data will ever be read.
        """

        read_size = max(self.CHUNK, size_hint)

        try:
            read = yield from self.loop.sock_recv(self.sock, read_size)
        except socket.error as error:
            raise SocketError(error)

        self.buf += read
        
        if read:
            log.debug("+%d/%d = %d", len(read), read_size, len(self.buf))
            return True
        else:
            log.debug("EOF @ %d + %d", len(self.buf), read_size)
            return False

    @asyncio.coroutine
    def read (self, size):
        """
            Read exactly the given size of data from the socket.
            
            Returns bytes.

            Raises EOFError.
        """

        while len(self.buf) < size:
            # guaranteed to make progress
            if not (yield from self._recv(size - len(self.buf))):
                raise EOFError()
        
        # split
        buf, self.buf = self.buf[:size], self.buf[size:]

        log.debug("%s -> %s + %s", size, len(buf), len(self.buf))

        return buf

    @asyncio.coroutine
    def readstr (self, size):
        """
            Read decoded string from socket.

            Raises EOFError.
        """

        buf = yield from self.read(size)

        try:
            line = buf.decode(self.encoding)
        except UnicodeDecodeError as error:
            log.warning("%s: %r", error, buf)
            raise
        
        return line

    @asyncio.coroutine
    def readline (self):
        """
            Read one line of input from the socket.

            Returns string without trailing newlines.

            Raises EOFError.
        """

        while self._readline not in self.buf:
            if not (yield from self._recv()):
                # XXX: accept truncated line at end?
                raise EOFError()
        
        # split
        buf, self.buf = self.buf.split(self._readline, 1)
        
        log.debug("%s", buf)

        # strip
        buf = buf.rstrip(self._readlines)
        
        try:
            line = buf.decode(self.encoding)
        except UnicodeDecodeError as error:
            log.warning("%s: %r", error, buf)
            raise StringError(error)
         
        return line

    @asyncio.coroutine
    def end (self):
        """
            Flush any pending buffers and shut down the socket.

            Discards any pending read buffer. Flushes any write buffers and sends FIN.
        """
        
        log.debug("%s: close", self)

        # XXX: doesn't this block with SO_LINGER somehow?
        self.sock.close()

    def __str__ (self):
        try:
            return self.peername()
        except OSError:
            return "<disconnected>"

class Server (qmsk.net.socket.Socket):
    """
        Trivial asyncio-enabled TCP listen server.
    """

    STREAM = Stream

    def __init__ (self, *args, **opts):
        super().__init__(*args, nonblocking=True, **opts)

    @asyncio.coroutine
    def accept (self, stream_cls=None, **opts):
        """
            Accept and return a single connection.

                stream_cls: Stream subclass to return
                **opts:     Stream(**opts)
        """

        if stream_cls is None:
            stream_cls = self.STREAM

        sock, addr = yield from self.loop.sock_accept(self.sock)

        log.debug("%s: accept: %s: %s", self, addr, sock)

        return stream_cls(sock, loop=self.loop, **opts)

    def __str__ (self):
        return self.sockname()

@asyncio.coroutine
def listen (port, host=None, loop=None, listen_backlog=LISTEN_BACKLOG, **opts):
    """
        Establish a Server() which accepts connections on a socket bound to the given addrinfo.
    """

    if loop is None:
        loop = asyncio.get_event_loop()

    # resolve
    try:
        ai = yield from loop.getaddrinfo(host, port, type=socket.SOCK_STREAM, flags=socket.AI_PASSIVE)
    except socket.gaierror as error:
        log.error("%s:%s: %s", host, port, error)
        raise

    sock = sockaddr = None

    for family, type, proto, canonname, sockaddr in ai:
        log.debug("%s:%s: family=%s type=%s sockaddr=%s", host, port, family, type, sockaddr)

        try:
            sock = socket.socket(family, type, proto)
        except socket.error as error:
            log.warning("%s:%s: socket(%s, %s, %s): %u", host, port, family, type, proto, error)

            sock = None

            continue

        try:
            sock.bind(sockaddr)
        except socket.error as error:
            log.warning("%s:%s: bind(%s): %s", host, port, sockaddr, error)

            sock.close()
            sock = None

            continue

    if not sock:
        raise Error("Unable to bind any listening socket")

    # out
    try:
        sock.listen(listen_backlog)
    except socket.error as error:
        log.warning("%s:%s: bind(%s): %s", host, port, sockaddr, error)
        raise

    log.debug("%s:%s: %s listening on %s", host, port, sock, sockaddr)

    return Server(sock, loop=loop)

@asyncio.coroutine
def connect (host, port, loop=None, **opts):
    """
        Establish a Stream() connected to a given remote host.
    """

    if loop is None:
        loop = asyncio.get_event_loop()
    
    # remote addr
    try:
        connect_ai = yield from loop.getaddrinfo(host, port, type=socket.SOCK_STREAM)
    except socket.gaierror as error:
        raise Error("%s:%s: %s" % (host, port, error))
    
    # TODO: except erorrs and try further alternatives, not just the first result?
    for family, type, proto, connect_canonname, connect_sockaddr in connect_ai:
        log.debug("socket family=%s type=%s: connect sockadr=%s", family, type, connect_sockaddr)
        
        try:
            sock = socket.socket(family, type, proto)
        except socket.error as error:
            raise Error("%s:%s: socket(family=%d, type=%d, proto=%d): %s" % (host, port, family, type, proto, errror))

        # we assume this is non-blocking for UDP sockets..
        try:
            sock.connect(connect_sockaddr)
        except socket.error as error:
            raise Error("%s:%s: connect(%s): %s" % (host, port, connect_sockaddr, error))

        return Stream(sock, loop=loop, **opts)
