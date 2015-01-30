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

class Index(qmsk.web.async.Handler):
    CLIENT = '/static/client/e2.html'

    def process(self):
        return werkzeug.redirect(self.CLIENT)

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

class HTMLBase(qmsk.web.html.HTMLMixin, BaseHandler):
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
        # TODO: 500 vs 400
        if self.error:
            return 400
        else:
            return 200

    def render_header(self):
        return html.div(id='header', class_='navbar')(
            html.div(class_='navbar-header')(
                html.a(href=self.url(Index), class_='navbar-brand')(self.TITLE),
            ),
            html.div(class_='narbar-collapse')(
                html.ul(class_='nav navbar-nav')(
                    html.li(class_=('active' if isinstance(self, page) else None))(
                        html.a(href=self.url(page))(page.PAGE_TITLE)
                    ) for page in HTML_PAGES
                ),
            ),
        )

    def render_status(self):
        return [ ]

    def render_content(self):
        raise NotImplementedError()

    def render(self):
        status = []

        for msg in self.render_status():
            status.append(html.p(msg))

        return html.div(class_='container-fluid', id='container')(
            self.render_header(),
            self.render_content(),
            html.div(id='status')(
                status or html.p("Ready")
            ),
        )

class HTMLPresets(HTMLBase):
    PAGE_TITLE = "Presets"

    def render_preset_destination(self, preset, destination):
        if preset == destination.program:
            format = "<{title}>"
        elif preset == destination.preview:
            format = "[{title}]"
        else:
            format = "({title})"

        return format.format(title=destination.title)

    def render_preset(self, preset):
        presets = self.app.presets
        css = set(['preset'])

        log.debug("preset=%s preview=%s program=%s", preset, presets.preview, presets.program)

        for destination in preset.destinations:
            if preset == destination.preview:
                css.add('preview')
            
            if preset == destination.program:
                css.add('program')

            if preset == presets.active:
                css.add('active')

        return html.button(
                type    = 'submit',
                name    = 'preset',
                value   = preset.preset,
                class_  = ' '.join(css) if css else None,
                id      = 'preset-{preset}'.format(preset=preset.preset),
                title   = ' + '.join(
                    self.render_preset_destination(preset, destination) for destination in preset.destinations
                ),
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
    def render_status(self):
        for value, message in (
                (self.preset, "Recalled preset {}"),
                (self.transition, "Transitioned {}"),
                (self.error, "Error: {}")
        ):
            if value is not None:
                yield message.format(value)

    def render_content(self):
        return html.form(action='', method='POST')(
            html.div(
                html.div(id='tools')(
                    html.button(type='submit', name='cut', value='cut', id='cut')("Cut"),
                    html.button(type='submit', name='autotrans', value='autotrans', id='autotrans')("Auto Trans"),
                ),
                html.div(id='presets', class_='presets')(
                    self.render_preset_group(group) for group in self.app.presets.groups
                ),
            ),
        )

class HTMLDestinations(HTMLBase):
    PAGE_TITLE = "Destinations"

    def render_destination_preset(self, preset, class_):
        css = set(['preset'])

        if preset:
            css.add(class_)
        else:
            css.add('empty')

        return html.button(class_=' '.join(css))(
            preset.title if preset else None
        )

    def render_destination(self, destination):
        return html.div(class_='destination')(
                html.h3(destination.title),
                self.render_destination_preset(destination.program, 'program'),
                self.render_destination_preset(destination.preview, 'preview'),
        )

    def render_content(self):
        return html.div(
            html.div(id='destinations', class_='presets')(
                self.render_destination(destination) for destination in self.app.presets.destinations
            ),
        )

HTML_PAGES = (
    HTMLPresets,
    HTMLDestinations,
)

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
    def process(self):
        self.presets = self.app.presets

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
                'presets': {preset.preset: self.render_preset(preset) for preset in self.presets},
                'groups': [self.render_group(group) for group in self.presets.groups],
                'destinations': [self.render_destination(destination) for destination in self.presets.destinations],
        }

class APIPreset(APIBase):
    def init(self):
        self.preset = None
        self.transition = self.error = None

    @asyncio.coroutine
    def process_async(self, preset=None):
        if preset:
            try:
                self.preset = self.app.presets[preset]
            except KeyError as error:
                raise werkzeug.exceptions.BadRequest("Invalid preset={preset}".format(preset=preset))

        post = self.request_post()

        log.info("content_type=%s, post=%r", self.request.mimetype, post)

        if post is not None:
            try:
                self.transition = yield from self.app.process(self.preset, post)
            except qmsk.e2.client.Error as error:
                self.error = werkzeug.exceptions.InternalServerError

    def render_json(self):
        out = { }

        if self.preset:
            out['preset'] = self.render_preset(self.preset)
        
        if self.transition is not None:
            out['transition'] = self.transition

        if self.error is not None:
            out['error'] = self.error

        return out

class E2Web(qmsk.web.async.Application):
    URLS = qmsk.web.urls.rules({
        '/':                            Index,
        '/presets':                     HTMLPresets,
        '/destinations':                HTMLDestinations,
        '/api/v1/':                     APIIndex,
        '/api/v1/preset/':              APIPreset,
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
    listen  = None,
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
            host    = listen,
            port    = port,
    )

    return application
