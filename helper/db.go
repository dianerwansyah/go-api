package helper

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
}

func LoadConfig() (Config, error) {
	var config Config
	configPath, err := filepath.Abs("devops/local/config.yaml")
	if err != nil {
		return config, errors.New("error getting absolute path for config file: " + err.Error())
	}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return config, errors.New("error reading config file: " + err.Error())
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return config, errors.New("error unmarshalling config file: " + err.Error())
	}

	return config, nil
}

func ConnectToPostgres(config Config, withDB bool) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
	)
	if withDB {
		connStr += fmt.Sprintf(" dbname=%s", config.Database.DBName)
	}
	return sql.Open("postgres", connStr)
}

func DatabaseExists(db *sql.DB, dbName string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbName)
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, errors.New("error checking if database exists: " + err.Error())
	}
	return exists, nil
}

func CreateDatabase(db *sql.DB, dbName string) error {
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return errors.New("error creating database: " + err.Error())
	}
	return nil
}

func SetupDatabase(config Config) (*sql.DB, error) {
	db, err := ConnectToPostgres(config, false)
	if err != nil {
		return nil, errors.New("error connecting to PostgreSQL: " + err.Error())
	}
	defer db.Close()

	dbExists, _ := DatabaseExists(db, config.Database.DBName)
	// if err != nil {
	// 	return nil, errors.New("error checking if database exists: " + err.Error())
	// }

	if !dbExists {
		err = CreateDatabase(db, config.Database.DBName)
		if err != nil {
			return nil, errors.New("error creating database: " + err.Error())
		}
	}

	dbWithDB, err := ConnectToPostgres(config, true)
	if err != nil {
		return nil, errors.New("error connecting to PostgreSQL with database: " + err.Error())
	}

	return dbWithDB, nil
}

// CreateTableFromModel creates a table in the database based on a model
func CreateTableFromModel(db *sql.DB, model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Struct {
		return errors.New("model is not a struct")
	}

	tableName := strings.ToLower(modelType.Name())

	var columns []string
	var primaryKey string
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		columnName := strings.ToLower(field.Name)
		columnType := getColumnType(field.Type)

		if columnType == "" {
			continue
		}

		if columnName == "id" {
			primaryKey = columnName
			columnType += " PRIMARY KEY"
		}

		// Check for unique constraint
		if strings.Contains(field.Tag.Get("key"), "uniq") {
			columnType += " UNIQUE"
		}

		columns = append(columns, fmt.Sprintf("%s %s", columnName, columnType))
	}

	if primaryKey == "" {
		return errors.New("primary key not defined in model")
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columns, ", "))

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error creating table %s: %v", tableName, err)
	}
	return nil
}

// UpdateTableFromModel updates a table in the database based on a model
func UpdateTableFromModel(db *sql.DB, model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Struct {
		return errors.New("model is not a struct")
	}

	tableName := strings.ToLower(modelType.Name())

	existingColumns, err := getExistingColumns(db, tableName)
	if err != nil {
		return err
	}

	var alterQueries []string
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		columnName := strings.ToLower(field.Name)
		columnType := getColumnType(field.Type)

		if columnType == "" {
			continue
		}

		if _, exists := existingColumns[columnName]; !exists {
			alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", tableName, columnName, columnType)
			alterQueries = append(alterQueries, alterQuery)

			// Check for unique constraint
			if strings.Contains(field.Tag.Get("key"), "uniq") {
				alterUniqueQuery := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s_unique UNIQUE (%s);", tableName, columnName, columnName)
				alterQueries = append(alterQueries, alterUniqueQuery)
			}
		}
	}

	for _, query := range alterQueries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing query %s: %v", query, err)
		}
	}
	return nil
}

// getColumnType returns the PostgreSQL column type for a given Go type
func getColumnType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int64:
		return "SERIAL"
	case reflect.String:
		return "TEXT"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "TIMESTAMP"
		}
	case reflect.Slice:
		// Handle many-to-many relationship by skipping
		return ""
	default:
		return ""
	}
	return ""
}

// getExistingColumns retrieves the existing columns of a table from the database
func getExistingColumns(db *sql.DB, tableName string) (map[string]string, error) {
	query := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s';", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying existing columns: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		columns[columnName] = dataType
	}

	return columns, nil
}

// CreateJoinTable creates a join table for many-to-many relationships
func CreateJoinTables(db *sql.DB, pairs [][]string) error {
	for _, tables := range pairs {
		if len(tables) < 2 {
			return errors.New("each entry must contain at least two table names")
		}

		joinTableName := strings.Join(tables, "_")
		var columns []string
		for _, table := range tables {
			column := fmt.Sprintf("%s_id INT REFERENCES %s(id)", table, table)
			columns = append(columns, column)
		}

		primaryKeys := make([]string, len(tables))
		for i, table := range tables {
			primaryKeys[i] = fmt.Sprintf("%s_id", table)
		}

		createJoinTableQuery := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				%s,
				PRIMARY KEY (%s)
			);
		`, joinTableName, strings.Join(columns, ", "), strings.Join(primaryKeys, ", "))

		//log.Println("Executing query:", createJoinTableQuery) // Log the query for debugging

		_, err := db.Exec(createJoinTableQuery)
		if err != nil {
			return fmt.Errorf("error creating join table %s: %v", joinTableName, err)
		}
		//log.Printf("Successfully created join table %s.\n", joinTableName)
	}
	return nil
}
