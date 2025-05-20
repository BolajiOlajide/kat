package runner

import "reflect"

func NonZeroPtr[T any](val T) *T {
	if reflect.ValueOf(val).IsZero() {
		return nil
	}
	return &val
}

var migrationLogsSchema = []dbSchema{
	{
		TableSchema: "public",
		TableName:   "migration_logs",
		ColumnName:  "id",
		DataType:    "bigint",
		IsNullable:  "NO",
	},
	{
		TableSchema: "public",
		TableName:   "migration_logs",
		ColumnName:  "name",
		DataType:    "text",
		IsNullable:  "NO",
	},
	{
		TableSchema:   "public",
		TableName:     "migration_logs",
		ColumnName:    "migration_time",
		DataType:      "timestamp with time zone",
		IsNullable:    "NO",
		ColumnDefault: NonZeroPtr("now()"),
	},
	{
		TableSchema: "public",
		TableName:   "migration_logs",
		ColumnName:  "duration",
		DataType:    "interval",
		IsNullable:  "NO",
	},
}

var usersSchema = dbSchema{
	TableSchema: "public",
	TableName:   "users",
	ColumnName:  "id",
	DataType:    "bigint",
	IsNullable:  "NO",
}

var productsSchema = dbSchema{
	TableSchema: "public",
	TableName:   "products",
	ColumnName:  "id",
	DataType:    "bigint",
	IsNullable:  "NO",
}

var rolesSchema = dbSchema{
	TableSchema: "public",
	TableName:   "roles",
	ColumnName:  "id",
	DataType:    "bigint",
	IsNullable:  "NO",
}

var transactionsSchema = dbSchema{
	TableSchema: "public",
	TableName:   "transactions",
	ColumnName:  "id",
	DataType:    "bigint",
	IsNullable:  "NO",
}
