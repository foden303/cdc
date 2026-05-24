package config

import "fmt"

// PostgresDSN builds a postgres connection string from common fields.
func PostgresDSN(host string, port int, username, password, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, database)
}

// MySQLDSN builds a mysql connection string from common fields.
func MySQLDSN(host string, port int, username, password, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
}
