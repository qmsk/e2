import asyncio
import aiohttp.wsgi
import logging; log = logging.getLogger('qmsk.e2.web')
import qmsk.web
import werkzeug

import qmsk.web.application
import qmsk.web.html
import qmsk.web.urls

class Index(qmsk.web.html.HTMLHandler):
    TITLE = "Hello World!!!!"

    def render(self):
        return self.html.h1(self.title())

class E2Web(qmsk.web.application.Application):
    URLS = qmsk.web.urls.rules({
        '/': Index,
    })

@asyncio.coroutine
def start (loop, port, host=None):
    application = E2Web()

    def server_factory():
        return aiohttp.wsgi.WSGIServerHttpProtocol(application,
                debug   = True,
        )

    server = yield from loop.create_server(server_factory,
            host    = host,
            port    = port,
    )

    return application
