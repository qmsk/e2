import logging; log = logging.getLogger('qmsk.e2.presets')
import os
import os.path
import dbm
import yaml

from xml.etree import ElementTree

class Error(Exception):
    pass

class XMLError(Error):
    pass

class Destination:
    def __init__ (self, index, *, title):
        self.index = index

        self.title = title

    def __lt__ (self, preset):
        return self.title < preset.title 
   
    def __str__ (self):
        return "{self.index}: {self.title}".format(self=self)

class Preset:
    def __init__ (self, preset, group, destinations, *, title):
        self.preset = preset
        self.group = group
        self.destinations = destinations

        self.title = title

    def __lt__ (self, preset):
        return self.title < preset.title 

    def __str__ (self):
        return "{self.preset}: {self.title} @ {self.group}".format(self=self)

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
    @classmethod
    def load (cls, xml_dir, yaml_file, db=None):
        obj = cls()

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

        if db:
            db = dbm.open(db, 'c')

            obj.load_db(db)

        return obj

    def __init__ (self):
        self._destinations = { }
        self.presets = { }

        self.default_group = PresetGroup(title=None)
        self._groups = { }

        # persistence
        self.db = None
        self.preview = { } # Destination.index: Preset
        self.program = { }

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
    
    def parse_xml_aux_dest (self, xml):
        return {
            'index': int(xml.find('OutCfgIndex').text),
            'title': xml.find('Name').text,
        }
        
    def parse_xml_screen_dest_index (self, xml):
        for xml_dest_out_map in xml.find('DestOutMapCol').findall('DestOutMap'):
            yield int(xml_dest_out_map.find('OutCfgIndex').text)

    def parse_xml_screen_dest (self, xml):
        return {
                'index': tuple(self.parse_xml_screen_dest_index(xml)),
                'title': xml.find('Name').text,
        }

    def load_xml_settings (self, xml):
        if xml.tag != 'System':
            raise XMLError("Unexpected preset root node: {xml}".format(xml=xml))

        xml_dest_mgr = xml.find('DestMgr')

        for xml_aux_dest_col in xml_dest_mgr.findall('AuxDestCol'):
            for xml_aux_dest in xml_aux_dest_col.findall('AuxDest'):
                self._load_destination(**self.parse_xml_aux_dest(xml_aux_dest))
        
        for xml_screen_dest_col in xml_dest_mgr.findall('ScreenDestCol'):
            for xml_screen_dest in xml_screen_dest_col.findall('ScreenDest'):
                self._load_destination(**self.parse_xml_screen_dest(xml_screen_dest))

    def parse_xml_preset (self, xml_preset):
        """
            Parse XML dump <Preset> and return { }
        """

        preset = int(xml_preset.attrib['id']) + 1
        title = xml_preset.find('Name').text
        destinations = []

        if '@' in title:
            title, group = title.split('@')
            title = title.strip()
            group = group.strip()
        else:
            group = None

        if group:
            group = self._load_group(group)
        
        for xml_screen_dest_col in xml_preset.findall('ScreenDestCol'):
            for xml_screen_dest in xml_screen_dest_col.findall('ScreenDest'):
                destinations.append(self._load_destination(**self.parse_xml_screen_dest(xml_screen_dest)))

        for xml_aux_dest_col in xml_preset.findall('AuxDestCol'):
            for xml_aux_dest in xml_aux_dest_col.findall('AuxDest'):
                destinations.append(self._load_destination(**self.parse_xml_aux_dest(xml_aux_dest)))

        return { 
                'preset': preset,
                'group': group,
                'destinations': destinations,
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

    def _load_destination (self, index, **item):
        obj = self._destinations.get(index)

        if obj is None:
            log.info("%s: %s", index, item)

            obj = self._destinations[index] = Destination(index, **item)

        return obj

    def _load_group (self, title):
        group = self._groups.get(title.lower())

        if group is None:
            group = self._groups[title.lower()] = PresetGroup(title=title)
        
        return group

    def _load_preset (self, preset, group=None, destinations=(), **item):
        """
            Load the given series of { 'preset': int, **opts } into (unique) Preset items.

                preset: int
                group: PresetGroup
        """

        log.info("%s @ %s: %s = %s", preset, group, item, ' + '.join(str(d) for d in destinations))

        if preset in self.presets:
            raise Error("Duplicate preset: {preset} = {item}".format(preset=preset, item=item))

        obj = self.presets[preset] = Preset(preset, group=group, destinations=destinations, **item)

        if not group:
            group = self.default_group

        group._add_preset(obj)

        return obj

    # state
    def _load_db_preset (self, *key):
        index = self.db.get('/'.join(str(k) for k in key))

        if index:
            return self.presets[int(index)]
        else:
            return None

    def _store_db_preset (self, preset, *key):
        self.db['/'.join(str(k) for k in key)] = str(preset.preset)

    def load_db (self, db):
        self.db = db

        for destination in self._destinations.values():
            destination.preview = self._load_db_preset('preview', destination.index)
            destination.program = self._load_db_preset('program', destination.index)

    def activate_preview (self, preset):
        self.preview = preset

        for destination in preset.destinations:
            log.info("%s: %s -> %s", destination, destination.preview, preset)

            destination.preview = preset

            self._store_db_preset(preset, 'preview', destination.index)
    
    def activate_program (self, preset=None):
        if preset is None:
            preset = self.preview
            self.preview = None
        
        self.program = preset

        for destination in preset.destinations:
            log.info("%s: %s -> %s", destination, destination.program, preset)

            destination.program = preset

            self._store_db_preset(preset, 'program', destination.index)

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

