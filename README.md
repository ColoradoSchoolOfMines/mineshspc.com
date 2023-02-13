# mineshspc.com

Source code for the [mineshspc.com](https://mineshspc.com) website.

![healthcheck](https://healthchecks.io/badge/fd6a8ec9-b3da-4bab-983a-183f2d/5Ll6vyEb-2/mineshspc.com.svg)

## Development Workflow

Install Go and [gow](https://github.com/mitranim/gow). Then run:
```
$ LOG_CONSOLE=1 gow -e=yaml,go,html,css run .
```
which will automatically restart the app whenever you make a change.

## Registration Flow

The following information needs to be gathered:

- For each teacher:

  - [x] Teacher name
  - [x] Teacher email

  - [x] School name
  - [x] School city
  - [x] School state

- For each team

  - [x] Team name
  - [x] Team division
  - [x] Team location (remote/in-person)
  - [x] Explanation of why the chosen division is correct

  - [x] Team members
    - [x] Name
    - [x] Age
    - [x] Email
    - [x] Parent's Email (if under 18)
    - [x] Have they participated in a previous competition?

- For each student when they confirm their email...

  Show them their Name, Email, Parent's Email, and whether they've participated
  before or not. (Show a thing saying they need to tell their teacher if they
  need to change something.)

  - [ ] Interest in campus tour
  - [ ] Dietary restrictions

- From students' parents (or student if old enough)

  - [ ] Competition waiver
  - [ ] Computer use waiver
  - [ ] Photo/multimedia release form

## License

The code is licensed under AGPLv3+. All of the content of the website (besides
the Colorado School of Mines logo) is Copyright (c) Mines ACM.
