#!/usr/bin/env python3
"""Template renderer for multiplan. Handles multiline substitution safely."""
import sys
import argparse

def render(template_path: str, vars: dict[str, str]) -> str:
    with open(template_path) as f:
        content = f.read()
    for key, value in vars.items():
        content = content.replace("{{" + key + "}}", value)
    return content

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("template")
    parser.add_argument("--var", action="append", nargs=2, metavar=("KEY", "VALUE"), default=[])
    parser.add_argument("--file", action="append", nargs=2, metavar=("KEY", "PATH"), default=[])
    args = parser.parse_args()

    vars = {}
    for key, value in args.var:
        vars[key] = value
    for key, path in args.file:
        with open(path) as f:
            vars[key] = f.read()

    print(render(args.template, vars))

if __name__ == "__main__":
    main()
