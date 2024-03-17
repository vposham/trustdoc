package dbtx

import (
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func getDocsMock() *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "doc_id", "title", "description", "file_name", "uploaded_by",
		"modified_at", "blockchain_hash", "uploaded_at", "last_updated_at"}).
		AddRow(1, "testDocId", "test document", "", "testdocument.pdf", 1, time.Time{}, "testhash",
			time.Time{}, time.Time{})
	return rows
}
