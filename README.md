# fishtank

A full-screen, monochrome text aquarium for Linux and macOS. Fish swim in
both directions while rooted plants wave. The tank fills the terminal and
automatically adjusts when it is resized.

[demo-video.webm](https://github.com/user-attachments/assets/cb585185-dd23-41cc-8f13-ea6b50ce6005)

## Build and run

Requires Go 1.24 or newer and a terminal with a valid `TERM` setting.
The first build downloads the [tcell](https://github.com/gdamore/tcell) terminal
library and its dependencies.

```sh
make build
./bin/fishtank
```

Or build and run in one step:

```sh
make run
```

Press **q**, **Escape**, or **Ctrl+C** to exit. The cursor and terminal settings
are restored on exit, including when receiving SIGINT or SIGTERM.

The artwork uses ASCII characters plus Unicode degree symbols for fish eyes,
and the terminal's default foreground and background colors. Rendering and fish
bounds use terminal display widths. Animation targets 20 frames per second. Fish and plant
populations scale with the available space, up to 24 fish and 40 plants.
Override either population with an exact, non-negative count:

```sh
./bin/fishtank --fish 30 --plants 12
# Or: make run ARGS="--fish 30 --plants 12"
```

Omitted options keep automatic scaling. Explicit counts remain fixed when the
terminal is resized and are not limited by the automatic caps; zero disables
that population. Use `--help` to list options.

Below 20 columns by 8 rows, the simulation pauses and shows a resize prompt;
quit keys remain active.

## Development

```sh
make test
make test-race
make vet
make fmt
```
