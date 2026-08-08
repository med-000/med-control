package sqlite

import "strings"

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func sqlNullableString(value string) string {
	if value == "" {
		return "NULL"
	}
	return sqlString(value)
}
