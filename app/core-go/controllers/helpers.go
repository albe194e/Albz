package controllers

import dsql "database/sql"

func nullStringValue(value dsql.NullString) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullString(value string) dsql.NullString {
	if value == "" {
		return dsql.NullString{}
	}

	return dsql.NullString{
		String: value,
		Valid:  true,
	}
}
