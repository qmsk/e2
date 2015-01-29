import asyncio
import aiohttp.wsgi
import logging; log = logging.getLogger('qmsk.e2.web')
import qmsk.web
import werkzeug

import qmsk.web.async
import qmsk.web.html
import qmsk.web.urls

html = qmsk.web.html.html5

class Index(qmsk.web.html.HTMLMixin, qmsk.web.async.Handler):
    TITLE = "Hello World!!!!"
    
    def init(self):
        self.preset = self.autotrans = self.error = None

    @asyncio.coroutine
    def process_async(self):
        if self.request.method == 'POST':
            self.preset, self.autotrans, self.error = yield from self.app.process(self.request.form)
        
    def render(self):
        return (
            html.h1(self.title()),
            html.p("Recalled preset {preset}".format(preset=self.preset)) if self.preset is not None else None,
            html.p("Autotransitioned {autotrans}".format(autotrans=self.autotrans)) if self.autotrans is not None else None,
            html.p("Error: {error}".format(error=self.error)) if self.error else None,
            html.form(action='', method='POST')(
                html.input(type='text', name='preset', placeholder='Preset number'),
                html.input(type='submit', name='cut', value='Cut'),
                html.input(type='submit', name='autotrans', value='Auto Trans'),
            )
        )

class E2Web(qmsk.web.async.Application):
    def __init__ (self, client):
        """
            client: qmsk.e2.client.E2Client
        """
        super().__init__()

        self.client = client
    
    @asyncio.coroutine
    def process(self, params):
        """
            Process an action request

            params: dict
                preset: int
                cut: *
                autotrans: *

            Returns preset, autotrans, error

            Raises werkzeug.HTTPException.
        """
        
        try:
            preset = params.get('preset', type=int)
        except ValueError as error:
            raise werkzeug.BadRequest("preset={preset}: {error}".format(preset=params.get('preset'), error=error))
        
        try:
            log.info("preset: %s", preset)

            if preset:
                yield from self.client.PRESET_recall(preset)

            if 'cut' in params:
                autotrans = 0
            elif 'autotrans' in params:
                autotrans = True
            else:
                autotrans = None 
            
            log.info("autotrans: %s", autotrans)

            if autotrans is not None:
                yield from self.client.ATRN(autotrans)

        except qmsk.e2.client.Error as error:
            return preset, autotrans, error
        else:
            return preset, autotrans, None

    URLS = qmsk.web.urls.rules({
        '/': Index,
    })

@asyncio.coroutine
def start (loop, client, port, host=None):
    """
        client: qmsk.e2.client.E2Client
    """

    application = E2Web(client)

    def server_factory():
        return aiohttp.wsgi.WSGIServerHttpProtocol(application,
                readpayload = True,
                debug       = True,
        )

    server = yield from loop.create_server(server_factory,
            host    = host,
            port    = port,
    )

    return application
