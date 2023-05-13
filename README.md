# fhd

The `fhd` “File History Data” package is for use by the GUI
[`FileHistory`](https://github.com/mark-summerfield/filehistory) application
and the command line [`fh`](https://github.com/mark-summerfield/fh) tool. It
provides the underlying functionality that these tools use.

`fhd` keeps all its data in a _single_ file with the `.fhd` extension.

## Data Structures

*TODO* replace with (generated) SVG

```
    "config" bucket
	key string # e.g., key="format" value=1
	value any  # e.g., key="ignore" value=bucket key=filename value=ε

    "states" bucket of StateItem (key + StateVal)
	key filename string
	value StateVal {
	    monitored bool
	    lastsid int # the most recent sid this file was saved into
	    filekind byte # so clients can see if they can offer diffs
	}

    "saveinfo" bucket of SaveInfoItem (key + SaveInfoVal)
	key sid int # > 0
	value SaveVal {
	    when time.Time
	    comment string
	}

    "saves" bucket of buckets
	key sid int # > 0
	value bucket
	    key filename string
	    value Entry
```

## License

Apache-2.0

---
