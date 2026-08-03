[![Check](https://github.com/etu/llr/actions/workflows/check.yml/badge.svg)](https://github.com/etu/llr/actions/workflows/check.yml)
[![Update](https://github.com/etu/llr/actions/workflows/update.yml/badge.svg)](https://github.com/etu/llr/actions/workflows/update.yml)

# `llr` - Line-length Limiter

The `llr` program reads a file or standard input, then it limits the length
of the lines printed to fit to your terminals width by truncating the lines.

```sh
llr [flags] [filename]
```

The following flags are available:
- `--width` or `-w`: specifies the maximum width of the output. If not
  specified, the default is the width of the terminal.
- `--debug` or `-d`: enables debug output.
- `--compact` or `-c`: compacts wide, space-aligned columns down to the
  width of their actual content before truncating. Many tools (such as
  `zfs list`) pad their first column to fit the widest *possible* value
  rather than the widest value actually present, which wastes space that
  could otherwise be used to show trailing columns.

It also accepts an argument filename to read from, if this filename isn't
specified or specified as `-`, it will read from standard input.

## Example usages

For example, to read the contents of a file named `input.txt` and print the
lines to standard output with a maximum width of 50 characters, you can use
the following command:

```sh
cat input.txt | llr -w 50
```

This may be more useful in programs that produce large and wide outputs where
you don't always care about the end of the lines, such example may in some
cases be output from =kubectl=.

`input-columns.txt` contains sample `zfs list` output whose `NAME` column is
padded much wider than any actual value. Compare truncating it as-is versus
compacting the columns first:

```sh
cat input-columns.txt | llr -w 60
cat input-columns.txt | llr -w 60 -c
```

Without `-c`, the wide `NAME` padding eats the whole width and every other
column is truncated away. With `-c`, the columns are shrunk to their real
content width first, so `USED`, `AVAIL`, `REFER`, and `MOUNTPOINT` all fit.

## Building

You can build the `llr` program by navigating to the project directory and
running the go build command:
```sh
go build -o llr
```

## Running tests

You can run the tests for the `llr` program by navigating to the project
directory and running the go test command:

```sh
go test
```
