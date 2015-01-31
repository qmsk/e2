import asyncio
import aiohttp.wsgi
import logging; log = logging.getLogger('qmsk.e2.web')
import qmsk.e2.client
import qmsk.e2.server
import qmsk.web.async
import qmsk.web.html
import qmsk.web.json
import qmsk.web.urls
import werkzeug
import werkzeug.exceptions

html = qmsk.web.html.html5

WEB_PORT = 8081
STATIC = './static'

class APIBase (qmsk.web.json.JSONMixin, qmsk.web.async.Handler):
    CORS_ORIGIN = '*'
    CORS_METHODS = ('GET', 'POST')
    CORS_HEADERS = ('Content-Type', 'Authorization')
    CORS_CREDENTIALS = True

    def render_preset(self, preset):
        destinations = dict()
        
        out = {
            'preset': preset.preset,
            'destinations': destinations,
            'title': preset.title,
            'group': preset.group.title if preset.group else None,
        }
       
        for destination in preset.destinations:
            if preset == destination.program:
                status = 'program'

            elif preset == destination.preview:
                status = 'preview'

            else:
                status = None

            destinations[destination.title] = status

            if status:
                out[status] = True

            if preset == self.app.presets.active:
                out['active'] = True
        
        return out

class APIIndex(APIBase):
    def init(self):
        self.presets = self.app.presets
        self.seq = self.app.server.seq

    def render_group (self, group):
        return {
                'title': group.title,
                'presets': [preset.preset for preset in group.presets],
        }

    def render_destination (self, destination):
        return {
                'outputs': destination.index,
                'title': destination.title,
                'preview': destination.preview.preset if destination.preview else None,
                'program': destination.program.preset if destination.program else None,
        }

    def render_json(self):
        return {
                'seq': self.seq,
                'presets': {preset.preset: self.render_preset(preset) for preset in self.presets},
                'groups': [self.render_group(group) for group in self.presets.groups],
                'destinations': [self.render_destination(destination) for destination in self.presets.destinations],
        }

class APIPreset(APIBase):
    """
        preset: Preset              activated preset, or requested preset, or active preset
        transition: True or int     activated transition
        seq: float                  current sequence number
    """

    def init(self):
        self.preset = None
        self.transition = self.error = None
        self.seq = self.app.server.seq

    @asyncio.coroutine
    def process_async(self, preset=None):
        """
            Raises werkzeug.exceptions.HTTPException.

                preset: int         - preset from URL
        """

        if preset:
            try:
                preset = self.app.presets[preset]
            except KeyError as error:
                raise werkzeug.exceptions.BadRequest("Invalid preset={preset}".format(preset=preset))
        else:
            preset = None
        
        post = self.request_post()

        if post is not None:
            try:
                self.preset, self.transition, self.seq = yield from self.app.process(preset, post)
            except qmsk.e2.server.SequenceError as error:
                raise werkzeug.exceptions.BadRequest(error)
            except qmsk.e2.client.Error as error:
                raise werkzeug.exceptions.InternalServerError(error)
            except qmsk.e2.server.Error as error:
                raise werkzeug.exceptions.InternalServerError(error)
        elif preset:
            self.preset = preset
        else:
            self.preset = self.app.presets.active

    def render_json(self):
        out = {
            'seq': self.seq,
        }

        if self.preset:
            out['preset'] = self.render_preset(self.preset)
        
        if self.transition is not None:
            out['transition'] = self.transition

        return out

class API(qmsk.web.async.Application):
    URLS = qmsk.web.urls.rules({
        '/v1/':                     APIIndex,
        '/v1/preset/':              APIPreset,
        '/v1/preset/<int:preset>':  APIPreset,
    })

    def __init__ (self, server):
        """
            server: qmsk.e2.server.Server
        """
        super().__init__()
        
        self.server = server
        self.presets = server.presets

    @asyncio.coroutine
    def process(self, preset, params):
        """
            Process an action request

            preset: Preset
            params: {
                cut: *
                autotrans: *
                transition: int
                seq: float or None
            }
        

            Raises qmsk.e2.client.Error, qmsk.e2.server.Error
        """

        if 'seq' in params:
            seq = float(params['seq'])
        else:
            seq = None

        if 'cut' in params:
            transition = 0
        elif 'autotrans' in params:
            transition = True
        elif 'transition' in params:
            transition = int(params['transition'])
        else:
            transition = None 

        active, seq = yield from self.server.activate(preset, transition, seq)
            
        return active, transition, seq

import argparse
import os.path

def parser (parser):
    group = parser.add_argument_group("qmsk.e2.web Options")
    group.add_argument('--e2-web-listen', metavar='ADDR',
        help="Web server listen address")
    group.add_argument('--e2-web-port', metavar='PORT', type=int, default=WEB_PORT,
        help="Web server port")
    group.add_argument('--e2-web-static', metavar='PATH', default=STATIC,
        help="Web server /static path")

@asyncio.coroutine
def apply (args, server, loop):
    """
        server: qmsk.e2.server.Server
    """
    
    # API
    api = API(server)
    
    # WSGI stack
    fallback = werkzeug.exceptions.NotFound()

    static = werkzeug.wsgi.SharedDataMiddleware(fallback, {
        '/':        os.path.join(args.e2_web_static, 'index.html'),
        '/lib':     os.path.join(args.e2_web_static, 'lib'),
        '/qmsk.e2': os.path.join(args.e2_web_static, 'qmsk.e2'),
    })

    application = werkzeug.wsgi.DispatcherMiddleware(static, {
        '/api': api,
    })
    
    # aiohttp Server
    def server_factory():
        return aiohttp.wsgi.WSGIServerHttpProtocol(application,
                readpayload = True,
                debug       = True,
        )

    server = yield from loop.create_server(server_factory,
            host    = args.e2_web_listen,
            port    = args.e2_web_port,
    )

    return application

