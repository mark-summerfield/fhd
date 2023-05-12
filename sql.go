// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

const (
	sqlPragmas = `
		PRAGMA encoding = 'UTF-8';
		PRAGMA foreign_keys = TRUE;
		PRAGMA synchronous = NORMAL;
		PRAGMA temp_store = MEMORY;`

	sqlCreateNames = `CREATE TABLE names (
		fid INTEGER NOT NULL,
		sid INTEGER NOT NULL,
		filename TEXT NOT NULL,

		PRIMARY KEY(fid, sid),
		FOREIGN KEY(fid) REFERENCES states(fid),
		FOREIGN KEY(sid) REFERENCES saves(sid),
		CHECK(fid > 0),
		CHECK(sid > 0),
		CHECK(LENGTH(filename) > 0)
	)`
	sqlCreateSaves = `CREATE TABLE saves (
		sid INTEGER NOT NULL PRIMARY KEY,
		fid INTEGER NOT NULL,
		sha TEXT NOT NULL,
		flag TEXT NOT NULL,
		data BLOB,

		FOREIGN KEY(fid) REFERENCES states(fid),
		CHECK(fid > 0),
		CHECK(sid > 0),
		CHECK(flag IN ('R', 'F', 'L'))
	)`
	sqlCreateSaveInfo = `CREATE TABLE saveinfo (
		sid INTEGER NOT NULL PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		comment TEXT,

		FOREIGN KEY(sid) REFERENCES saves(sid),
		CHECK(sid > 0)
	)`
	sqlCreateStates = `CREATE TABLE states (
		fid INTEGER NOT NULL PRIMARY KEY,
		monitored TEXT NOT NULL,
		lastsid INTEGER NOT NULL,
		lastkind TEXT NOT NULL,

		CHECK(fid > 0),
		CHECK(monitored IN ('Y', 'N')),
		CHECK(lastsid > 0),
		CHECK(lastkind IN ('B', 'I', 'T'))
	)`

	sqlGetNameForFid = `SELECT filename FROM (
		SELECT filename, sid FROM names WHERE fid = :fid)
		WHERE sid = MAX(sid)`
)
