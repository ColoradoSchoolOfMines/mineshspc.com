# mineshspc.com

Source code for the [mineshspc.com](https://mineshspc.com) website.

![healthcheck](https://healthchecks.io/badge/fd6a8ec9-b3da-4bab-983a-183f2d/5Ll6vyEb-2/mineshspc.com.svg)

## Development Workflow

Install Go and [gow](https://github.com/mitranim/gow). Then run:
```
$ LOG_CONSOLE=1 gow -e=yaml,go,html,css run .
```
which will automatically restart the app whenever you make a change.

## License

The code is licensed under AGPLv3+. All of the content of the website (besides
the Colorado School of Mines logo) is Copyright (c) Mines ACM 2023.
