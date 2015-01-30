import asyncio
import autobahn.asyncio.websocket
import logging; log = logging.getLogger('qmsk.e2.websocket')

class WebSocket(autobahn.asyncio.websocket.WebSocketServerProtocol):
    def onConnect(self, request):
        self.peer = request.peer

        log.info("%s: %s", self, request)

    def onOpen(self):
        log.info("%s", self)

        self.factory.clients.add(self)

    def update(self):
        log.info("%s", self)
        self.sendMessage('update'.encode('utf-8'))

    def onClose(self, wasClean, code, reason):
        log.info("%s: %s", self, reason)
        self.factory.clients.remove(self)

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

@asyncio.coroutine
def start (presets,
    loop,
    port,
    listen  = None,
    static  = None,
):
    """
        client: qmsk.e2.client.E2Client
    """

    factory = WebSocketServer(presets)

    server = yield from loop.create_server(factory,
            host    = listen,
            port    = port,
    )

    return factory
