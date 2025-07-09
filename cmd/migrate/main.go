package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bekhuli/cli_migrator/db"
)

type Field struct {
	Name string
	Type string
}

func goTypeToPostgres(t string) string {
	switch strings.ToLower(t) {
	case "string":
		return "TEXT"
	case "int", "integer":
		return "INTEGER"
	case "smallint":
		return "SMALLINT"
	case "datetime":
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func parseFields(raw string) []Field {
	parts := strings.Split(raw, ",")
	fields := []Field{}
	for _, p := range parts {
		pair := strings.Split(p, ":")
		if len(pair) < 2 {
			continue
		}
		fields = append(fields, Field{
			Name: pair[0],
			Type: goTypeToPostgres(pair[1]),
		})
	}
	return fields
}

func generateSQL(table string, fields []Field) string {
	sql := fmt.Sprintf("CREATE TABLE %s (\n", table)
	for _, f := range fields {
		sql += fmt.Sprintf("  %s %s,\n", f.Name, f.Type)
	}
	sql += "  id SERIAL PRIMARY KEY\n);\n"
	return sql
}

func writeSQLFile(filename string, content string) error {
	return os.WriteFile("migrations/"+filename, []byte(content), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go table_name --fields=\"name:string,age:int\"")
		return
	}

	tableName := os.Args[1]

	fieldFlag := flag.NewFlagSet("fields", flag.ExitOnError)
	fieldsStr := fieldFlag.String("fields", "", "comma-separated fields: name:type")
	fieldFlag.Parse(os.Args[2:])

	if *fieldsStr == "" {
		fmt.Println("No fields provided")
		return
	}

	fields := parseFields(*fieldsStr)
	sql := generateSQL(tableName, fields)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_create_%s.sql", timestamp, tableName)
	os.MkdirAll("migrations", 0755)
	if err := writeSQLFile(filename, sql); err != nil {
		fmt.Println("Failed to write SQL file:", err)
		return
	}

	fmt.Println("SQL file generated:", filename)

	db.Connect()
	_, err := db.DB.Exec(sql)
	if err != nil {
		fmt.Println("Failed to apply migration:", err)
		return
	}
	fmt.Println("Migration applied to database")
}
