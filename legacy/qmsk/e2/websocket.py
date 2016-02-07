import asyncio
import autobahn.asyncio.websocket
import logging; log = logging.getLogger('qmsk.e2.websocket')

WEBSOCKET_PORT = 8082

class WebSocket(autobahn.asyncio.websocket.WebSocketServerProtocol):
    def onConnect(self, request):
        self.peer = request.peer

        log.info("%s", self)

    def onOpen(self):
        log.info("%s", self)

        self.factory.clients.add(self)

    def update(self):
        log.info("%s", self)
        self.sendMessage('update'.encode('utf-8'))

    def onClose(self, wasClean, code, reason):
        if self in self.factory.clients:
                log.info("%s: %s", self, reason)
                self.factory.clients.remove(self)
        else:
                log.warning("%s: unknown client close: %s", self, reason)

    def __str__(self):
        return str(self.peer)

class WebSocketServer(autobahn.asyncio.websocket.WebSocketServerFactory):
    protocol = WebSocket

    def __init__(self, presets, **opts):
        super().__init__(**opts)

        self.presets = presets

        self.clients = set()
        
        self.presets.add_notify(self.update)

    def update(self):
        for client in self.clients:
            client.update()

    def stop(self):
        self.presets.del_notify(self.update)

import argparse

def parser (parser):
    group = parser.add_argument_group("qmsk.e2.websocket Options")
    group.add_argument('--e2-websocket-port', metavar='PORT', type=int, default=WEBSOCKET_PORT,
        help="WebSocket server port")

@asyncio.coroutine
def apply (args, presets, loop):
    """
        presets: qmsk.e2.presets.E2Presets
    """

    factory = WebSocketServer(presets)

    server = yield from loop.create_server(factory,
            host    = args.e2_web_listen,
            port    = args.e2_websocket_port,
    )

    return factory
