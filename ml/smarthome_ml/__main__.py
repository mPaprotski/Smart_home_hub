"""Contract scaffold CLI; this is not a running inference service."""

import argparse

from . import __version__


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Observation contract scaffold; inference unavailable"
    )
    parser.add_argument("--version", action="version", version=__version__)
    parser.parse_args()
    parser.exit(2, "NOT_IMPLEMENTED: ML/CV/DSP runtime is a future development stage\n")


if __name__ == "__main__":
    main()
