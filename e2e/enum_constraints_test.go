//go:build integration

package e2e

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

const defaultDatabaseDSN = "root:root@tcp(127.0.0.1:3306)/espresso-api?parseTime=true"

func TestMySQLEnumCheckConstraints(t *testing.T) {
	databaseDSN := os.Getenv("E2E_DATABASE_DSN")
	if databaseDSN == "" {
		databaseDSN = defaultDatabaseDSN
	}

	db, err := dbsql.Open("mysql", databaseDSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	testID := time.Now().UnixNano()
	roasterID := insertRow(t, ctx, db, "INSERT INTO roasters (name) VALUES (?)", fmt.Sprintf("enum-test-roaster-%d", testID))
	sheetID := insertRow(t, ctx, db, "INSERT INTO sheets (name) VALUES (?)", fmt.Sprintf("enum-test-sheet-%d", testID))
	beanID := insertRow(t, ctx, db, "INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)", fmt.Sprintf("enum-test-beans-%d", testID), roasterID, nil, 2)

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM beans WHERE id = ?", beanID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM sheets WHERE id = ?", sheetID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM roasters WHERE id = ?", roasterID)
	})

	assertCheckConstraint(t, ctx, db, "chk_beans_roast_level", "INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)", fmt.Sprintf("enum-test-invalid-beans-%d", testID), roasterID, nil, 5)
	assertCheckConstraint(t, ctx, db, "chk_shots_comparison_with_previous_result", "INSERT INTO shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sheetID, beanID, 12, 18, 36, 24, 93, 8, false, false, 4, "invalid comparison")
}

func insertRow(t *testing.T, ctx context.Context, db *dbsql.DB, query string, args ...any) int64 {
	t.Helper()

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		t.Fatalf("insert row: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get inserted row id: %v", err)
	}

	return id
}

func assertCheckConstraint(t *testing.T, ctx context.Context, db *dbsql.DB, constraint, query string, args ...any) {
	t.Helper()

	_, err := db.ExecContext(ctx, query, args...)
	if err == nil {
		t.Fatalf("query succeeded; expected check constraint %q to reject it", constraint)
	}

	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		t.Fatalf("error = %T %v, want MySQLError", err, err)
	}
	if mysqlErr.Number != 3819 {
		t.Errorf("MySQL error number = %d, want 3819", mysqlErr.Number)
	}
	if !strings.Contains(mysqlErr.Message, constraint) {
		t.Errorf("MySQL error message = %q, want constraint %q", mysqlErr.Message, constraint)
	}
}
