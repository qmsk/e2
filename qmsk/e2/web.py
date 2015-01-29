import asyncio
import aiohttp.wsgi
import logging; log = logging.getLogger('qmsk.e2.web')
import qmsk.web
import werkzeug

import qmsk.web.async
import qmsk.web.html
import qmsk.web.json
import qmsk.web.urls

html = qmsk.web.html.html5

class BaseHandler(qmsk.web.async.Handler):
    def init(self):
        self.preset = self.autotrans = self.error = None

    @asyncio.coroutine
    def process_async(self):
        if self.request.method == 'POST':
            self.preset, self.autotrans, self.error = yield from self.app.process(self.request.form)
 
class Index(qmsk.web.html.HTMLMixin, BaseHandler):
    TITLE = "Encore2 Control"
       

    def render_preset(self, preset):
        return html.button(type='submit', name='preset', value=preset.preset)(preset.title)

    def render(self):
        return (
            html.h1(self.title()),
            html.form(action='', method='POST')(
                html.div(
                    self.render_preset(preset) for preset in self.app.presets
                ),
                html.div(
                    html.input(type='submit', name='cut', value='Cut'),
                    html.input(type='submit', name='autotrans', value='Auto Trans'),
                )
            ),
            html.div(
                html.p("Recalled preset {preset}".format(preset=self.preset)) if self.preset is not None else None,
                html.p("Autotransitioned {autotrans}".format(autotrans=self.autotrans)) if self.autotrans is not None else None,
                html.p("Error: {error}".format(error=self.error)) if self.error else None,
            ),
        )

class API(qmsk.web.json.JSONMixin, BaseHandler):
    def render_json(self):
        out = dict()
        
        if self.preset is not None:
            out['preset'] = self.preset

        if self.autotrans is not None:
            out['autotrans'] = self.autotrans

        if self.error is not None:
            out['error'] = self.error

        return out

class E2Web(qmsk.web.async.Application):
    URLS = qmsk.web.urls.rules({
        '/':        Index,
        '/api/v1':  API,
    })

    def __init__ (self, client, presets):
        """
            client: qmsk.e2.client.E2Client
        """
        super().__init__()

        self.client = client
        self.presets = presets
    
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

@asyncio.coroutine
def start (client, presets, loop, port, host=None):
    """
        client: qmsk.e2.client.E2Client
    """

    application = E2Web(client, presets)

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
