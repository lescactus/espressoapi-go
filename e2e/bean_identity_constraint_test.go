//go:build integration

package e2e

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBeanIdentityConstraint(t *testing.T) {
	config := testDatabaseConfig(t)
	db, err := dbsql.Open(config.driverName, config.datasource)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	testID := time.Now().UnixNano()
	roasterID := testRowID(testID, 20)
	datedBeanID := testRowID(testID, 21)
	duplicateDatedBeanID := testRowID(testID, 22)
	otherDateBeanID := testRowID(testID, 23)
	undatedBeanID := testRowID(testID, 24)
	duplicateUndatedBeanID := testRowID(testID, 25)
	name := fmt.Sprintf("identity-test-beans-%d", testID)
	roastDate := time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)
	otherRoastDate := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)

	insertRow(t, ctx, db, config, roasterID,
		"INSERT INTO roasters (id, name) VALUES (?, ?)",
		roasterID, fmt.Sprintf("identity-test-roaster-%d", testID))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM beans WHERE roaster_id = ?"), roasterID)
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM roasters WHERE id = ?"), roasterID)
	})

	// Dated identity: exact duplicate rejected, different date accepted.
	insertRow(t, ctx, db, config, datedBeanID,
		"INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)",
		datedBeanID, name, roasterID, roastDate, 2)
	_, err = db.ExecContext(ctx, config.bind(
		"INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)"),
		duplicateDatedBeanID, name, roasterID, roastDate, 2)
	assertUniqueConstraintError(t, config, err, "uq_beans_identity")
	insertRow(t, ctx, db, config, otherDateBeanID,
		"INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)",
		otherDateBeanID, name, roasterID, otherRoastDate, 2)

	// NULL identity: NULL matches NULL, so the second undated bean is rejected.
	insertRow(t, ctx, db, config, undatedBeanID,
		"INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)",
		undatedBeanID, name, roasterID, nil, 2)
	_, err = db.ExecContext(ctx, config.bind(
		"INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)"),
		duplicateUndatedBeanID, name, roasterID, nil, 2)
	assertUniqueConstraintError(t, config, err, "uq_beans_identity")
}

func assertUniqueConstraintError(t *testing.T, config databaseConfig, err error, constraint string) {
	t.Helper()
	if err == nil {
		t.Fatalf("insert succeeded; expected unique constraint %q to reject it", constraint)
	}

	switch config.typeName {
	case "mysql":
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) {
			t.Fatalf("error = %T %v, want MySQLError", err, err)
		}
		if mysqlErr.Number != 1062 {
			t.Errorf("MySQL error number = %d, want 1062", mysqlErr.Number)
		}
		if !strings.Contains(mysqlErr.Message, constraint) {
			t.Errorf("MySQL error message = %q, want constraint %q", mysqlErr.Message, constraint)
		}
	case "postgres":
		var postgresErr *pgconn.PgError
		if !errors.As(err, &postgresErr) {
			t.Fatalf("error = %T %v, want PgError", err, err)
		}
		if postgresErr.Code != "23505" {
			t.Errorf("PostgreSQL error code = %q, want 23505", postgresErr.Code)
		}
		if postgresErr.ConstraintName != constraint {
			t.Errorf("PostgreSQL constraint = %q, want %q", postgresErr.ConstraintName, constraint)
		}
	}
}
