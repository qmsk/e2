import logging; log = logging.getLogger('qmsk.e2.presets')
import yaml

class Preset:
    def __init__ (self, preset, *, title):
        self.preset = preset
        self.title = title

    def __str__ (self):
        return "{self.preset}: {self.title}".format(self=self)

class E2Presets:
    @classmethod
    def load_yaml (cls, file):
        data = yaml.safe_load(file)

        log.debug("%s: %s")

        return cls(**data)

    def __init__ (self, presets):
        self.presets = {id: Preset(id, **values) for id, values in presets.items()}

        self.preview = self.program = None

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
