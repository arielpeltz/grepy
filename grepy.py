#!/usr/bin/env python
import sys
import argparse
import re


def color(filename, ln,  mtchs):
    line = mtchs[0].string
    last = 0
    out = "{file} ({ln}) ".format(file=filename, ln=ln)
    for m in mtchs:
        out += "{nomatch}{highlight}{match}{default}".format(nomatch=line[last:m.start()],
                                                                   match=line[m.start():m.end()],
                                                                   highlight='\033[32m', default='\33[39m')
        last = m.end()
    print(out)


def machine(filename, ln, mtchs):
    print("{file}:{ln}:{line}".format(file=filename, ln=ln, line=mtchs[0].string.strip()))


def underline(filename, ln, mtchs):
    line = mtchs[0].string
    out1 = "{file} ({ln}) ".format(file=filename, ln=ln)
    out2 = ' ' * len(out1)
    last = 0
    for m in mtchs:
        out2 += ' ' * (m.start()-last)
        out2 += '^' * (m.end()-m.start())
        last = m.end()
    print("{0}{1}{2}".format(out1, line, out2))


def parse_args():
    parser = argparse.ArgumentParser("grepy")
    parser.add_argument('regex', type=re.compile, help='Regex to match in files')
    parser.add_argument('files', nargs='*', type=argparse.FileType('r'), default=[sys.stdin],
                        help='Files to look in, if not provided will read from stdin')
    fmt = parser.add_mutually_exclusive_group()
    fmt.add_argument('-c', '--color', dest='format', action='store_const', const=color,
                     help='Highlights the matches in color')
    fmt.add_argument('-u', '--underline', dest='format', action='store_const', const=underline,
                     help='Highlight the matches with \'^\' under the line')
    fmt.add_argument('-m', '--machine', dest='format', action='store_const', const=machine,
                     help='Print in machine format [file name]:[line number]:[match text]')
    parser.set_defaults(format=machine)
    return parser.parse_args()


if __name__ == '__main__':
    args = parse_args()
    for infile in args.files:
        line_no = 0
        for line in infile:
            line_no += 1
            m = list(args.regex.finditer(line))
            if len(m) > 0:
                args.format(infile.name, line_no, m)
