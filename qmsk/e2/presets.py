import logging; log = logging.getLogger('qmsk.e2.presets')
import shelve
import yaml

class Error(Exception):
    pass

class Preset:
    def __init__ (self, preset, *, title):
        self.preset = preset

        self.title = title

    def __eq__ (self, preset):
        return isinstance(preset, Preset) and preset.preset == self.preset

    def __str__ (self):
        return "{self.preset}: {self.title}".format(self=self)

class PresetGroup:
    def __init__ (self, presets, *, title):
        self.presets = presets
        
        self.title = title

class DBProperty:
    def __init__ (self, name):
        self.name = name

    def __get__ (self, obj, type=None):
        log.debug("%s", self.name)

        return obj.db.get(self.name)

    def __set__ (self, obj, value):
        log.debug("%s: %s", self.name, value)

        obj.db[self.name] = value

    def __del__ (self, obj):
        log.debug("%s", self.name)

        del obj.db[self.name]

class E2Presets:
    preview = DBProperty('preview')
    program = DBProperty('program')

    @classmethod
    def load_yaml (cls, file, db=None):
        data = yaml.safe_load(file)

        if db:
            db = shelve.open(db, 'c')
        else:
            db = None

        log.debug("data=%s, db=%s", file, data, db)

        return cls(db, **data)

    def __init__ (self, db, presets={ }, groups=[]):
        self.db = db
        self.groups = [ ]
        self.presets = { }

        self.groups.append(PresetGroup(list(self._init_presets(presets)), title=None))
        
        for item in groups:
            presets = list(self._init_presets(item.pop('presets')))

            group = PresetGroup(presets, **item)

            self.groups.append(group)
        
        if db is None:
            # no presistence
            self.preview = None
            self.program = None

    def _init_presets (self, presets):
        for item in presets:
            id = item.pop('preset')

            if id in self.presets:
                raise Error("Duplicate preset: {id} = {item}".format(id=id, item=item))

            preset = self.presets[id] = Preset(id, **item)

            yield preset

    def activate_preview (self, preset):
        log.info("%s -> %s", self.preview, preset)
        self.preview = preset
    
    def activate_program (self, preset=None):
        if preset is None:
            preset = self.preview
            self.preview = None

        log.info("%s -> %s", self.program, preset)
        self.program = preset

    def __iter__ (self):
        for preset in self.presets.values():
            yield preset

    def __getitem__ (self, key):
        return self.presets[key]

    def close(self):
        if self.db:
            self.db.close()

