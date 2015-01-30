from distutils.core import setup
from glob import glob

def globs (*patterns):
    return [path for pattern in patterns for path in glob(pattern)]

__version__ = '0.9'

setup(
    name        = 'qmsk-e2',
    version     = __version__,
    description = "Encore2 Web Preset Manager",
    url         = 'https://github.com/SpComb/qmsk-e2',

    namespace_packages  = [ 'qmsk' ],
    packages    = [
        'qmsk.e2',
        'qmsk.net',
    ],
    py_modules  = [
        'qmsk.cli',
    ],

    scripts     = [
        'qmsk-e2-client',
        'qmsk-e2-web',
    ],

    data_files  = [
        ('static/qmsk.e2', glob('static/qmsk.e2/*')),
        ('static/client', glob('static/client/*')),
    ],
)
