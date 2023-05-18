# fhd

The `fhd` “File History Data” library is for use by the GUI
[`FileHistory`](https://github.com/mark-summerfield/filehistory) application
and the command line [`fh`](https://github.com/mark-summerfield/fh) tool. It
provides the underlying functionality that these tools use.

> The File History project is an **experiment** — _use at your own risk!_
> The purpose is to provide file history for users who don't have the
> technical knowledge/interest in using a full-blown version control system,
> even finding GUIs like `tortoisehg` and `qgit` too complicated to use,
> but who want to have something more convenient than saving copies of their
> files with dates added to the filenames or similar.

## Data Structures

The `fhd` library keeps all the data in a _single_ file with the `.fhd`
extension. This data is in the form of a key–value store with some values
nested key–values in their own right.

![The `fhd` Key–Value Data Store](diag/db.svg)

The `config` bucket's `format` value is the `.fhd` file format number. And
the `config` bucket's `ignore` value is a bucket whose keys are filenames
or globs to be ignored and whose values are empty.

The `states` bucket holds the current state. The `LastSid` is the most
recent `SID` the corresponding file was saved into. The `FileKind` is `B`
(binary), `I` (image), or `T` (text): useful for clients to see if they can
offer diffs.

## License

Apache-2.0

---
