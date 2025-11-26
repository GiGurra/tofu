# tail

Output the last part of files

## Synopsis

`tofu tail [flags] [files...]`

## Description

Print the last 10 lines of each FILE to standard output. With more than one FILE, precede each with a header giving the file name.

## Options

*   `-n, --lines int`: Output the last N lines, instead of the last 10 (default 10)
*   `-f, --follow`: Output appended data as the file grows
*   `-q, --quiet`: Never output headers giving file names
*   `-v, --verbose`: Always output headers giving file names

## Examples

    tofu tail file.txt
    tofu tail -n 20 file.txt
    tofu tail -f /var/log/syslog
