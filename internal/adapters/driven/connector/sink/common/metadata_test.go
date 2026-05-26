package common

import "testing"

func TestPostgresTableKeyDefaultsSchema(t *testing.T) {
	key, meta, err := PostgresTableKey("", "users")
	if err != nil {
		t.Fatal(err)
	}
	if key != "public.users" {
		t.Fatalf("key = %q", key)
	}
	if meta.Schema != "public" || meta.Table != "users" {
		t.Fatalf("meta = %+v", meta)
	}
}

func TestPostgresTableKeyUsesQualifiedTable(t *testing.T) {
	key, meta, err := PostgresTableKey("", "tenant.users")
	if err != nil {
		t.Fatal(err)
	}
	if key != "tenant.users" {
		t.Fatalf("key = %q", key)
	}
	if meta.Schema != "tenant" || meta.Table != "users" {
		t.Fatalf("meta = %+v", meta)
	}
}

func TestMySQLTableKeyUsesDatabase(t *testing.T) {
	key, meta, err := MySQLTableKey("app", "users")
	if err != nil {
		t.Fatal(err)
	}
	if key != "app.users" {
		t.Fatalf("key = %q", key)
	}
	if meta.Schema != "app" || meta.Table != "users" {
		t.Fatalf("meta = %+v", meta)
	}
}

func TestTableKeyRejectsEmptyTable(t *testing.T) {
	if _, _, err := PostgresTableKey("", ""); err == nil {
		t.Fatal("expected error")
	}
	if _, _, err := MySQLTableKey("", ""); err == nil {
		t.Fatal("expected error")
	}
}
