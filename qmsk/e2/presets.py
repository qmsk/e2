import logging; log = logging.getLogger('qmsk.e2.presets')
import os
import os.path
import shelve
import yaml

from xml.etree import ElementTree

class Error(Exception):
    pass

class XMLError(Error):
    pass

class Preset:
    def __init__ (self, preset, *, title):
        self.preset = preset

        self.title = title

    def __eq__ (self, preset):
        return isinstance(preset, Preset) and preset.preset == self.preset
    
    def __lt__ (self, preset):
        return self.title < preset.title 

    def __str__ (self):
        return "{self.preset}: {self.title}".format(self=self)

class PresetGroup:
    def __init__ (self, *, title):
        self._presets = []
        
        self.title = title

    def _add_preset (self, preset):
        self._presets.append(preset)

    @property
    def presets (self):
        return tuple(sorted(self._presets))

    def __str__ (self):
        return self.title

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
    def load (cls, xml_dir, yaml_file, db=None):
        if db:
            db = shelve.open(db, 'c')
        else:
            db = None

        obj = cls(db)

        if xml_dir:
            xml_settings = os.path.join(xml_dir, 'settings_backup.xml')

            log.info("%s", xml_settings)

            obj.load_xml_settings(ElementTree.parse(xml_settings).getroot())
            
            xml_presets = os.path.join(xml_dir, 'presets')

            for name in os.listdir(xml_presets):
                xml_preset = os.path.join(xml_presets, name)

                log.info("%s", xml_preset)

                obj.load_xml_preset(ElementTree.parse(xml_preset).getroot())
        
        if yaml_file:
            obj.load_yaml(**yaml.safe_load(yaml_file))

        return obj

    def __init__ (self, db):
        self.db = db
        self.presets = { }

        self.default_group = PresetGroup(title=None)
        self._groups = { }

        if db is None:
            # no presistence
            self.preview = None
            self.program = None

    def load_yaml (self, presets={ }, groups=[]):
        """
            Load user-editable metadata from the YAML object attributes given as keyword arguments.
        """

        for preset, item in presets.items():
            self._load_presets(preset, group=None, **item)
        
        for item in groups:
            presets = item.pop('presets')

            group = self._load_group(**item)
            
            for item in presets:
                self._load_preset(group=group, **item)
    
    def load_xml_settings (self, xml):
        pass

    def parse_xml_preset (self, xml_preset):
        """
            Parse XML dump <Preset> and return { }
        """

        preset = int(xml_preset.attrib['id']) + 1
        title = xml_preset.find('Name').text

        if '@' in title:
            title, group = title.split('@')
            title = title.strip()
            group = group.strip()
        else:
            group = None

        if group:
            group = self._load_group(group)

        return { 
                'preset': preset,
                'group': group,
                'title': title,
        }

    def load_xml_preset (self, xml):
        """
            Load an XML dump <PresetMgr> root element and load the <Preset>s
        """

        if xml.tag != 'PresetMgr':
            raise XMLError("Unexpected preset root node: {xml}".format(xml=xml))
        
        for xml_preset in xml.findall('Preset'):
            preset = self._load_preset(**self.parse_xml_preset(xml_preset))

    def _load_group (self, title):
        group = self._groups.get(title.lower())

        if group is None:
            group = self._groups[title.lower()] = PresetGroup(title=title)
        
        return group

    def _load_preset (self, preset, group=None, **opts):
        """
            Load the given series of { 'preset': int, **opts } into (unique) Preset items.

                preset: int
                group: PresetGroup
        """

        log.info("%s @ %s: %s", preset, group, opts)

        if preset in self.presets:
            raise Error("Duplicate preset: {preset} = {item}".format(preset=preset, item=item))

        obj = self.presets[preset] = Preset(preset, **opts)

        if not group:
            group = self.default_group

        group._add_preset(obj)

        return obj

    def activate_preview (self, preset):
        log.info("%s -> %s", self.preview, preset)
        self.preview = preset
    
    def activate_program (self, preset=None):
        if preset is None:
            preset = self.preview
            self.preview = None

        log.info("%s -> %s", self.program, preset)
        self.program = preset

    @property
    def groups (self):
        yield self.default_group

        for name, group in sorted(self._groups.items()):
            yield group

    def __iter__ (self):
        for preset in self.presets.values():
            yield preset

    def __getitem__ (self, key):
        return self.presets[key]

    def close(self):
        if self.db:
            self.db.close()

