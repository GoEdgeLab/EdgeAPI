package setup

import (
	"encoding/json"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestSQLDump_Dump(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})

	dump := NewSQLDump()
	result, err := dump.Dump(db)
	if err != nil {
		t.Fatal(err)
	}

	// Table
	for _, table := range result.Tables {
		_ = table
		//t.Log(table.Name, table.Engine, table.Charset)

		/**for _, field := range table.Fields {
			t.Log("===", field.Name, ":", field.Definition)
		}**/
		/**for _, index := range table.Indexes {
			t.Log("===", index.Name, ":", index.Definition)
		}**/

		/**for _, record := range table.Records {
			t.Log(record.Id, record.Values)
		}**/
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(data), "bytes")
}

func TestSQLDump_Apply(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})

	dump := NewSQLDump()
	result, err := dump.Dump(db)
	if err != nil {
		t.Fatal(err)
	}

	db2, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	ops, err := dump.Apply(db2, result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
	if len(ops) > 0 {
		for _, op := range ops {
			t.Log("", op)
		}
	}
}
