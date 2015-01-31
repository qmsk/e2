import argparse
import asyncio
import logging; log = logging.getLogger('qmsk.e2.server')
import qmsk.net.tcp
import qmsk.e2.client
import qmsk.e2.presets
import qmsk.e2.web
import qmsk.e2.websocket
import signal
import time

class Error(Exception):
    pass

class SequenceError(Error):
    pass

class Server:
    def __init__ (self, loop):
        self.loop = loop

        self.client = None
        self.presets = None
        
        self.seq = time.time()
        self.lock = asyncio.Lock()

    @asyncio.coroutine
    def start (self, args):
        try:
            self.client = yield from qmsk.e2.client.apply(args)
        except qmsk.net.tcp.Error as error:
            log.error("%s: failed to connect: %s", args.e2_host, error)
            return 1

        self.presets = qmsk.e2.presets.apply(args)

        if not self.presets.presets:
            log.error("no presets given")
            return 1
        
        self.web = yield from qmsk.e2.web.apply(args, self, self.loop)

        self.websocket = yield from qmsk.e2.websocket.apply(args, self.presets, self.loop)

    @asyncio.coroutine
    def activate(self, preset=None, transition=None, seq=None):
        """
            Activate presets.
                preset: qmsk.e2.presets.Preset      activate given preset, placing it on preview for its destinations
                transition: True or int             transition active preset to program on its destinations
                seq: float or None                  serialize activate()'s between clients.

            Returns active:Preset, seq:float

            Raises qmsk.e2.client.Error, SequenceError.
        """

        log.debug("enter")
        
        with (yield from self.lock):
            active = self.presets.active
            
            # seq
            if seq is None:
                pass
            elif seq != self.seq:
                raise SequenceError("Sequence mismatch: {} != {}".format(seq, self.seq))
            else:
                pass

            new_seq = self.seq = time.time()

            log.info("seq %s -> %s", seq, self.seq)
           
            # preset -> preview?
            if preset:
                log.info("preset: %s", preset)

                yield from self.client.PRESET_recall(preset.preset)
                
                active = self.presets.activate_preview(preset)

            # preview -> program?
            if transition is not None:
                log.info("%s: transition %s", active, transition)

                yield from self.client.ATRN(transition)
                
                active = self.presets.activate_program()
            
            log.debug("preset=%s transition=%s seq=%s -> active=%s seq=%s", preset, transition, seq, active, new_seq)
        
            log.debug("exit")
            
        return active, new_seq

    @asyncio.coroutine
    def stop (self):
        self.presets.close()
        self.websocket.stop()

def signal_stop (server):
    """
        Stop event loop gracefully on signal.
    """

    loop = asyncio.get_event_loop()

    # tell our Server to stop
    log.info("signalling server stop...")
    do_stop = asyncio.async(server.stop(), loop=loop)

    def _stop (do_stop):
        log.info("stopping event loop...")

        # this will ignore any new callbacks added, and cause the loop.run_forever() in main to return
        loop.stop()

    # once stopped, wind down the event loop
    do_stop.add_done_callback(_stop)

def main (argv):
    parser = qmsk.cli.parser()

    qmsk.e2.client.parser(parser)
    qmsk.e2.presets.parser(parser)
    qmsk.e2.web.parser(parser)
    qmsk.e2.websocket.parser(parser)
    
    # setup
    args = qmsk.cli.parse(parser, argv)
    loop = asyncio.get_event_loop()

    server = Server(loop)

    # start
    do_start = asyncio.async(server.start(args))

    try:
        log.info("start event loop")
        ret = loop.run_until_complete(do_start)

    except Exception as error:
        log.exception("Failed to start")
        return 1

    else:
        if ret:
            return ret

        log.info("startup complete")

    # run
    loop.add_signal_handler(signal.SIGINT, signal_stop, server)

    try:
        log.info("enter event loop")
        loop.run_forever()

    except Exception as error:
        log.exception("Failed to start")
        return 1

    else:
        log.info("exit")
        return 0
