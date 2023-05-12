// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	sqlPragmas = `
		PRAGMA encoding = 'UTF-8';
		PRAGMA foreign_keys = TRUE;
		PRAGMA synchronous = NORMAL;
		PRAGMA temp_store = MEMORY;`

	sqlCreateSids = `CREATE TABLE sids (
		sid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		comment TEXT,

		CHECK(sid > 0)
	)`
	sqlCreateFids = `CREATE TABLE fids (
		fid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,

		CHECK(fid > 0)
	)`
	sqlCreateSaves = `CREATE TABLE saves (
		sid INTEGER NOT NULL,
		fid INTEGER NOT NULL,
		sha TEXT NOT NULL,
		flag TEXT NOT NULL,
		data BLOB,

		PRIMARY KEY(fid, sid),
		FOREIGN KEY(sid) REFERENCES sids(fid),
		FOREIGN KEY(fid) REFERENCES fids(fid),
		CHECK(fid > 0),
		CHECK(sid > 0),
		CHECK(flag IN ('R', 'F', 'L'))
	)`
	sqlCreateStates = `CREATE TABLE states (
		fid INTEGER NOT NULL PRIMARY KEY,
		filename TEXT NOT NULL,
		monitored TEXT NOT NULL,
		lastsid INTEGER NOT NULL,
		lastkind TEXT NOT NULL,

		FOREIGN KEY(fid) REFERENCES fids(fid),
		CHECK(fid > 0),
		CHECK(monitored IN ('Y', 'N')),
		CHECK(lastsid > 0),
		CHECK(lastkind IN ('B', 'I', 'T'))
	)`
)
