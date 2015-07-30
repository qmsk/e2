import asyncio
import logging; log = logging.getLogger('qmsk.e2.client')
import qmsk.net.tcp

class Error(Exception):
    pass

class CommandError(Error):
    pass

class E2Client:
    """
        safe:bool   only send safe commands (preview-only), noop program commands
    """

    PORT = 9878

    @classmethod
    @asyncio.coroutine
    def connect (cls, host, port=PORT, **opts):
        """
            Raises qmsk.net.tcp.Error
        """

        stream = yield from qmsk.net.tcp.connect(host, port)
        
        log.info("%s: connected: %s", host, stream)

        return cls(stream, **opts)

    def __init__ (self, stream, safe=None):
        self.stream = stream

        self.safe = safe # safe mode
        
        # only one command at any time
        self.lock = asyncio.Lock()

    @asyncio.coroutine
    def cmd (self, cmd, *args, safe=False):
        """
            Raises qmsk.net.tcp.Error, CommandError
        """

        if self.safe and not safe:
            log.warn("%s: noop unsafe", cmd)
            return
 
        # XXX: implement timeouts to ensure livelyness
        with (yield from self.lock):
            line = ' '.join([cmd] + list(args))

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
            preset:         int 0-1000, or str '%d.%d'

            Raises ValueError, CommandError, qmsk.net.tcp.Error.
        """

        if isinstance(preset, int):
            preset = '%d' % preset
        elif all(c in '0123456789.' for c in preset):
            preset = preset
        else:
            raise ValueError(preset)

        yield from self.cmd('PRESET', '-r', preset, safe=True)

    @asyncio.coroutine
    def ATRN (self, transTime=True):
        """
            transTime:int   frames or True
            
            Raises ValueError, CommandError, qmsk.net.tcp.Error.
        """
       
        if transTime is True:
            yield from self.cmd('ATRN')
        elif isinstance(transTime, int):
            yield from self.cmd('ATRN', str(transTime))
        else:
            raise ValueError(preset)

    def __str__ (self):
        return str(self.stream)

# cli
import argparse

def parser (parser):
    group = parser.add_argument_group("qmsk.e2.client Options")
    group.add_argument('--e2-host', metavar='HOST',
        help="Encore2 host address")
    group.add_argument('--e2-safe', action='store_true',
        help="Encore2 safe mode")
    
@asyncio.coroutine
def apply (args):
    client = yield from E2Client.connect(args.e2_host,
            safe    = args.e2_safe,
    )

    return client

