import logging; log = logging.getLogger('qmsk.e2.presets')
import yaml

class Error(Exception):
    pass

class Preset:
    def __init__ (self, preset, *, title):
        self.preset = preset

        self.title = title

    def __str__ (self):
        return "{self.preset}: {self.title}".format(self=self)

class PresetGroup:
    def __init__ (self, presets, *, title):
        self.presets = presets
        
        self.title = title

class E2Presets:
    @classmethod
    def load_yaml (cls, file):
        data = yaml.safe_load(file)

        log.debug("%s: %s")

        return cls(**data)

    def __init__ (self, presets={ }, groups=[]):
        self.groups = [ ]
        self.presets = { }

        self.groups.append(PresetGroup(list(self._init_presets(presets)), title=None))
        
        for item in groups:
            presets = list(self._init_presets(item.pop('presets')))

            group = PresetGroup(presets, **item)

            self.groups.append(group)

        self.preview = self.program = None

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
