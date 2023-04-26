# fhd

The `fhd` “File History Data” package is for use by the GUI
[`FileHistory`](https://github.com/mark-summerfield/filehistory) application
and the command line [`fh`](https://github.com/mark-summerfield/fh) tool. It
provides the underlying functionality that these tools use.

The `fhd` package allows its clients to:

- maintain a monitoring list of one or more files in a folder (and
  optionally in one or more subfolders) to be captured (i.e., add or
  delete files from the monitoring list;
- capture the monitored files as they are at this moment (in VCS-speak:
  check in)—each capture may include an optional comment and optional tag; 
- rename one or more files (so history is preserved despite the name change)
- view previous captures of any monitored file;
- compare any previous capture of any monitored file with the current
  file or any other previous capture.
- restore any previous capture of any monitored file;
- purge a monitored file and its entire history so it disappears forever.

`fhd` keeps all its data in a _single_ file with the `.fhd` extension.

## License

Apache-2.0

---
