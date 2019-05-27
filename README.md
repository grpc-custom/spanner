# spanner

```proto
message Singer {
  option (spanner.ddl.schema) = {
    database: "test"
    table: "Singers"
    index {
      name: "SingersByFirstLastName"
      columns: "FirstName, LastName"
    }
    index {
      name: "SingersByFirstLastNameNoNulls"
      columns: "FirstName, LastName"
      null_filtered: true
    }
    primary_key: "SingerId"
  };
  int64  id          = 1
    [(spanner.ddl.column) = {name: "SingerId"}];
  string first_name  = 2
    [(spanner.ddl.column) = {size: 1024}];
  string last_name   = 3
    [(spanner.ddl.column) = {size: 1024 nullable: true}];
  bytes  singer_info = 4
    [(spanner.ddl.column) = {nullable: true}];
}

message Album {
  option (spanner.ddl.schema) = {
    database: "test"
    table: "Albums"
    index {
      name: "AlbumsByAlbumTitle"
      columns: "AlbumTitle"
    }
    index {
      name: "AlbumsByUniqueAlbumId"
      columns: "AlbumId"
      unique: true
    }
    primary_key: "SingerId, AlbumId"
    interleave: "Singers"
  };
  int64  id               = 1
    [(spanner.ddl.column) = {name: "SingerId"}];
  int64  album_id         = 2;
  string album_title      = 3
    [(spanner.ddl.column) = {nullable: true}];
  int64  marketing_budget = 4
    [(spanner.ddl.column) = {nullable: true}];
}

message Song {
  option (spanner.ddl.schema) = {
    database: "test"
    table: "Songs"
    primary_key: "SingerId, AlbumId, TrackId"
    interleave: "Albums"
    index {
      name: "SongsBySongName"
      columns: "SongName"
    }
    index {
      name: "SongsBySingerAlbumSongName"
      columns: "SingerId, AlbumId, SongName"
      interleave: "Albums"
    }
    index {
      name: "SongsBySingerAlbumSongNameDesc"
      columns: "SingerId, AlbumId, SongName DESC"
      interleave: "Albums"
    }
  };
  int64 id         = 1
    [(spanner.ddl.column) = {name: "SingerId"}];
  int64  album_id  = 2;
  int64  track_id  = 3;
  string song_name = 4
    [(spanner.ddl.column) = {size: MAX, nullable: true}];
}
```

test時に設定させる
```go
func TestMain(m *testing.M) {
	opt := WithPrefixDatabase("")
	Setup()
	code := m.Run()
	os.Exit(code)
}
```
