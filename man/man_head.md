# head

Output the first part of files

## Synopsis

`tofu head [flags] [files...]`

## Description

Print the first 10 lines of each FILE to standard output. With more than one FILE, precede each with a header giving the file name.

## Options

*   `-n, --lines int`: Output the first N lines, instead of the first 10 (default 10)
*   `-q, --quiet`: Never output headers giving file names
*   `-v, --verbose`: Always output headers giving file names

## Examples

    tofu head file.txt
    tofu head -n 20 file.txt
    tofu head file1.txt file2.txt
