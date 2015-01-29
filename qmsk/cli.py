import argparse
import logging; log = logging.getLogger('qmsk.args')
import sys

def parser (**opts):
    parser = argparse.ArgumentParser(**opts)

    args = parser.add_argument_group("Generic options")
    args.add_argument('-q', '--quiet',      dest='log_level', action='store_const', const=logging.ERROR,
            help="Less output")
    args.add_argument('-v', '--verbose',    dest='log_level', action='store_const', const=logging.INFO,
            help="More output")
    args.add_argument('-D', '--debug-all',  dest='log_level', action='store_const', const=logging.DEBUG,
            help="Most output")

    args.add_argument('-d', '--debug-module',   dest='log_debug', action='append',
            help="Debugging output for given module")

    parser.set_defaults(
            log_level       = logging.WARNING,
            log_debug       = [ ],
    )

    return parser

def parse (parser, argv):
    """
        Parse given sys.argv using the ArgumentParser returned by parser()
    """

    args = parser.parse_args(argv[1:])

    logging.basicConfig(
            format      = "{levelname:<8} {name:>30}:{funcName:<20}: {message}",
            style       = '{',
            level       = args.log_level,
    )

    for logger in args.log_debug:
        logging.getLogger(logger).setLevel(logging.DEBUG)
    
    log.debug(args)

    return args

def main (main):
    """
        Run given main(argv) which returns exit status.
    """

    sys.exit(main(sys.argv))
