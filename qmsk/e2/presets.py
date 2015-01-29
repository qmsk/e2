import logging; log = logging.getLogger('qmsk.e2.presets')
import yaml

class E2Presets:
    @classmethod
    def load_yaml (cls, file):
        data = yaml.safe_load(file)

        log.debug("%s: %s")

        return cls(**data)

    def __init__ (self, presets):
        self.presets = presets


