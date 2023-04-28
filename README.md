# fhd

The `fhd` “File History Data” package is for use by the GUI
[`FileHistory`](https://github.com/mark-summerfield/filehistory) application
and the command line [`fh`](https://github.com/mark-summerfield/fh) tool. It
provides the underlying functionality that these tools use.

The `fhd` package allows its clients to:

- monitor one or more files in the current folder (and optionally
  subfolders);
- set one or more files to be ignored;
- stop (or restart) monitoring a monitored file;
- state - this should return a list of the files `fhd` knows about and their
  state and whether any have been saved, and a list of any files in the same
  folder(s) whose state is unknown — these need to be either monitored or
  ignored;
- save a snapshot of the monitored files (“check in”) with an optional
  comment;
- extract any previously saved copy (using a unique generated filename or
  a given filename);
- delete one or more previously saved copies of a monitored file (thus
  forgetting parts of its history);
- purge a monitored file and its entire history so it disappears forever.

`fhd` keeps all its data in a _single_ file with the `.fhd` extension.

## License

Apache-2.0

---
