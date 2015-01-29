import asyncio
import logging; log = logging.getLogger('qmsk.e2.client')
import qmsk.net.tcp

class Error(Exception):
    pass

class CommandError(Error):
    pass

class E2Client:
    PORT = 9878

    @classmethod
    @asyncio.coroutine
    def connect (cls, host, port=PORT):
        """
            Raises qmsk.net.tcp.Error
        """

        stream = yield from qmsk.net.tcp.connect(host, port)
        
        log.info("%s: connected: %s", host, stream)

        return cls(stream)

    def __init__ (self, stream):
        self.stream = stream
        
        # only one command at any time
        self.lock = asyncio.Lock()

    @asyncio.coroutine
    def cmd (self, cmd, *args):
        """
            Raises qmsk.net.tcp.Error, CommandError
        """
        
        # XXX: implement timeouts to ensure livelyness
        with (yield from self.lock):
            line = ' '.join([cmd] + [str(arg) for arg in args])

            log.info("%s: %s", self, line)
            
            yield from self.stream.writeline(line)

            while True:
                line = yield from self.stream.readline()

                log.debug("%s: %s: %r", self, cmd, line)

                if line.startswith('\x04'):
                    # wtf
                    line = line[1:]

                parts = line.split()

                if not parts:
                    # skip
                    continue

                elif len(parts) == 3 and parts[0] == cmd and parts[1] == '-e':
                    try:
                        err = int(parts[2])
                    except ValueError as error:
                        raise CommandError("%s: invalid error status: %s: %s", cmd, parts[2], line)

                    break
                else:
                    log.warning("%s: %s: %r", self, cmd, line)
               
            if err:
                raise CommandError(cmd, err)

    @asyncio.coroutine
    def PRESET_recall (self, preset):
        """
            preset:int      0-1000 
        """

        yield from self.cmd('PRESET', '-r', preset)

    @asyncio.coroutine
    def ATRN (self, transTime=True):
        """
            transTime:int   frames or True
        """
        
        if transTime is True:
            yield from self.cmd('ATRN')
        else:
            yield from self.cmd('ATRN', transTime)

    def __str__ (self):
        return str(self.stream)
