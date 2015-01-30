import asyncio
import aiohttp.wsgi
import logging; log = logging.getLogger('qmsk.e2.web')
import qmsk.web
import time
import werkzeug
import werkzeug.exceptions

import qmsk.web.async
import qmsk.web.html
import qmsk.web.json
import qmsk.web.urls

html = qmsk.web.html.html5

class BaseHandler(qmsk.web.async.Handler):
    def init(self):
        self.preset = self.transition = self.error = None

    @asyncio.coroutine
    def process_async(self):
        try:
            preset = self.request.form.get('preset', type=int)
        except ValueError as error:
            self.error = error
            return
        
        if preset:
            try:
                self.preset = self.app.presets[preset]
            except KeyError as error:
                self.error = error
                return

        if self.request.method == 'POST':
            try:
                self.transition = yield from self.app.process(self.preset, self.request.form)
            except qmsk.e2.client.Error as error:
                self.error = error
                return
 
class Index(qmsk.web.html.HTMLMixin, BaseHandler):
    TITLE = "Encore2 Control"

    CSS = (
        'https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css',
        'https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap-theme.min.css',
        
        # de-cache
        '/static/qmsk.e2/e2.css?' + str(time.time()),
    )

    JS = (
        '//code.jquery.com/jquery-1.11.2.min.js',
        'https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/js/bootstrap.min.js',
    )

    HEAD = (
        html.meta(name="viewport", content="width=device-width, initial-scale=1"),
    )

    def status(self):
        if self.error:
            return 400
        else:
            return 200

    def render_preset(self, preset):
        presets = self.app.presets
        css = set(['preset'])

        log.info("preset=%s preview=%s program=%s", preset, presets.preview, presets.program)

        if preset == presets.preview:
            css.add('preview')
        
        if preset == presets.program:
            css.add('program')

        return html.button(
                type    = 'submit',
                name    = 'preset',
                value   = preset.preset,
                class_  = ' '.join(css) if css else None,
                id      = 'preset-{preset}'.format(preset=preset.preset)
        )(preset.title)

    def render_preset_group (self, group):
        if not group.presets:
            return

        return html.div(class_='preset-group')(
                html.h3(group.title) if group.title else None,
                [
                    self.render_preset(preset) for preset in group.presets
                ],
        )

    def render(self):
        status = []

        for value, message in (
                (self.preset, "Recalled preset {}"),
                (self.transition, "Transitioned {}"),
                (self.error, "Error: {}")
        ):
            if value is not None:
                status.append(html.p(message.format(value)))

        return html.div(class_='container-fluid', id='container')(
            html.div(
                html.div(id='header')(
                    html.h1(self.title()),
                ),
            ),
            html.form(action='', method='POST')(
                html.div(
                    html.div(id='tools')(
                        html.button(type='submit', name='cut', value='cut', id='cut')("Cut"),
                        html.button(type='submit', name='autotrans', value='autotrans', id='autotrans')("Auto Trans"),
                    ),
                    html.div(id='presets')(
                        self.render_preset_group(group) for group in self.app.presets.groups
                    ),
                ),
                html.div(id='status')(
                    status or html.p("Ready")
                ),
            ),
        )

class APIBase (qmsk.web.json.JSONMixin, qmsk.web.async.Handler):
    CORS_ORIGIN = '*'
    CORS_METHODS = ('GET', 'POST')

    def render_preset(self, preset):
        presets = self.app.presets

        out = {
            'preset': preset.preset,
            'title': preset.title,
        }

        if preset == presets.preview:
            out['preview'] = True

        if preset == presets.program:
            out['program'] = True

        return out

class APIIndex(APIBase):
    def process(self):
        self.presets = self.app.presets

    def render_group (self, group):
        return {
                'title': group.title,
                'presets': [preset.preset for preset in group.presets],
        }

    def render_json(self):
        return {
                'presets': {preset.preset: self.render_preset(preset) for preset in self.presets},
                'groups': [self.render_group(group) for group in self.presets.groups],
        }

class APIPreset(APIBase):
    def init(self):
        self.transition = self.error = None

    @asyncio.coroutine
    def process_async(self, preset):
        try:
            self.preset = self.app.presets[preset]
        except KeyError as error:
            raise werkzeug.exceptions.BadRequest("Invalid preset={preset}".format(preset=preset))

        if self.request.method == 'POST':
            try:
                self.transition = yield from self.app.process(self.preset, self.request.form)
            except qmsk.e2.client.Error as error:
                self.error = werkzeug.exceptions.InternalServerError

    def render_json(self):
        out = self.render_preset(self.preset)
        
        if self.transition is not None:
            out['transition'] = self.transition

        if self.error is not None:
            out['error'] = self.error

        return out

class E2Web(qmsk.web.async.Application):
    URLS = qmsk.web.urls.rules({
        '/':                            Index,
        '/api/v1/':                     APIIndex,
        '/api/v1/preset/<int:preset>':  APIPreset,
    })

    def __init__ (self, client, presets):
        """
            client: qmsk.e2.client.E2Client
        """
        super().__init__()

        self.client = client
        self.presets = presets
    
    @asyncio.coroutine
    def process(self, preset, params):
        """
            Process an action request

            params: dict
                preset: int
                cut: *
                autotrans: *

            Returns transition value.

            Raises qmsk.e2.client.Error
        """
       
        # preset -> preview?
        log.info("preset: %s", preset)

        if preset:
            yield from self.client.PRESET_recall(preset)
            
            self.presets.activate_preview(preset)

        # preview -> program?
        if 'cut' in params:
            transition = 0
        elif 'autotrans' in params:
            transition = True
        elif 'transition' in params:
            transition = int(params['transition'])
        else:
            transition = None 
        
        log.info("transition: %s", transition)

        if transition is not None:
            yield from self.client.ATRN(transition)
            
            self.presets.activate_program()

        return transition

@asyncio.coroutine
def start (client, presets, loop, port,
    host    = None,
    static  = None,
):
    """
        client: qmsk.e2.client.E2Client
    """

    application = E2Web(client, presets)

    if static:
        application = werkzeug.wsgi.SharedDataMiddleware(application, static)

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
