# Nordwand - rolling hash

This is a binary file diffing algorithm using a rolling checksum to find blocks
with a fixed length in two files/byte slices.

This needs go 1.22. Run the tests with

``` shell
go test ./...
```

I didn't optimize the code for execution speed. Some of the requirements have
been unclear to me:

- The code operates on byte slices, which would break with large enough
  files. But this could easily be changed by taking an `io.Reader` as input
  arguments.
- The 'delta' only contains the locations to be read from the updated file, not
  the actual data. If we did add the actual data, we may run out of memory with
  large enough files.
