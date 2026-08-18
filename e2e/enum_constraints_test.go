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
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/repository"
	mysqlroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/roaster"
	postgresroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/roaster"
)

const (
	defaultMySQLDatabaseDSN    = "root:root@tcp(127.0.0.1:3306)/espresso-api?parseTime=true"
	defaultPostgresDatabaseDSN = "postgres://root:root@127.0.0.1:5432/espresso-api?sslmode=disable"
)

type databaseConfig struct {
	typeName   string
	driverName string
	datasource string
}

func TestEnumCheckConstraints(t *testing.T) {
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
	roasterID := testRowID(testID, 0)
	sheetID := testRowID(testID, 1)
	beanID := testRowID(testID, 2)
	invalidBeansID := testRowID(testID, 3)
	invalidShotID := testRowID(testID, 4)
	insertRow(t, ctx, db, config, roasterID, "INSERT INTO roasters (id, name) VALUES (?, ?)", roasterID, fmt.Sprintf("enum-test-roaster-%d", testID))
	insertRow(t, ctx, db, config, sheetID, "INSERT INTO sheets (id, name) VALUES (?, ?)", sheetID, fmt.Sprintf("enum-test-sheet-%d", testID))
	insertRow(t, ctx, db, config, beanID, "INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)", beanID, fmt.Sprintf("enum-test-beans-%d", testID), roasterID, nil, 2)

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM beans WHERE id = ?"), beanID)
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM sheets WHERE id = ?"), sheetID)
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM roasters WHERE id = ?"), roasterID)
	})

	assertCheckConstraint(t, ctx, db, config, "chk_beans_roast_level", "INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)", invalidBeansID, fmt.Sprintf("enum-test-invalid-beans-%d", testID), roasterID, nil, 5)
	assertCheckConstraint(t, ctx, db, config, "chk_shots_comparison_with_previous_result", "INSERT INTO shots (id, sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", invalidShotID, sheetID, beanID, 12, 18, 36, 24, 93, 8, false, false, 4, "invalid comparison")
}

func TestForeignKeyDeleteConstraint(t *testing.T) {
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
	roasterID := testRowID(testID, 10)
	beanID := testRowID(testID, 11)
	insertRow(t, ctx, db, config, roasterID, "INSERT INTO roasters (id, name) VALUES (?, ?)", roasterID, fmt.Sprintf("foreign-key-test-roaster-%d", testID))
	insertRow(t, ctx, db, config, beanID, "INSERT INTO beans (id, name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?, ?)", beanID, fmt.Sprintf("foreign-key-test-beans-%d", testID), roasterID, nil, 2)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM beans WHERE id = ?"), beanID)
		_, _ = db.ExecContext(context.Background(), config.bind("DELETE FROM roasters WHERE id = ?"), roasterID)
	})

	var roasterRepository repository.RoasterRepository
	sqlxdb := sqlx.NewDb(db, config.driverName)
	switch config.typeName {
	case "mysql":
		roasterRepository = mysqlroaster.New(sqlxdb)
	case "postgres":
		roasterRepository = postgresroaster.New(sqlxdb)
	default:
		t.Fatalf("unsupported DATABASE_TYPE %q", config.typeName)
	}

	err = roasterRepository.DeleteRoasterById(ctx, int(roasterID))
	if !errors.Is(err, domainerrors.ErrBeansForeignKeyConstraint) {
		t.Errorf("DeleteRoasterById() error = %v, want %v", err, domainerrors.ErrBeansForeignKeyConstraint)
	}
}

func testDatabaseConfig(t *testing.T) databaseConfig {
	t.Helper()

	typeName := os.Getenv("DATABASE_TYPE")
	if typeName == "" {
		typeName = "mysql"
	}

	datasource := os.Getenv("E2E_DATABASE_DSN")
	switch typeName {
	case "mysql":
		if datasource == "" {
			datasource = defaultMySQLDatabaseDSN
		}
		return databaseConfig{typeName: typeName, driverName: "mysql", datasource: datasource}
	case "postgres":
		if datasource == "" {
			datasource = defaultPostgresDatabaseDSN
		}
		return databaseConfig{typeName: typeName, driverName: "pgx", datasource: datasource}
	default:
		t.Fatalf("unsupported DATABASE_TYPE %q", typeName)
		return databaseConfig{}
	}
}

func (config databaseConfig) bind(query string) string {
	if config.typeName == "postgres" {
		return sqlx.Rebind(sqlx.DOLLAR, query)
	}

	return query
}

func testRowID(testID, offset int64) int64 {
	return -((testID % 1_000_000_000) + offset + 1)
}

func insertRow(t *testing.T, ctx context.Context, db *dbsql.DB, config databaseConfig, id int64, query string, args ...any) {
	t.Helper()

	if config.typeName == "postgres" {
		var insertedID int64
		if err := db.QueryRowContext(ctx, config.bind(query)+" RETURNING id", args...).Scan(&insertedID); err != nil {
			t.Fatalf("insert row: %v", err)
		}
		if insertedID != id {
			t.Fatalf("inserted id = %d, want %d", insertedID, id)
		}
		return
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("insert row: %v", err)
	}
}

func assertCheckConstraint(t *testing.T, ctx context.Context, db *dbsql.DB, config databaseConfig, constraint, query string, args ...any) {
	t.Helper()

	_, err := db.ExecContext(ctx, config.bind(query), args...)
	if err == nil {
		t.Fatalf("query succeeded; expected check constraint %q to reject it", constraint)
	}

	switch config.typeName {
	case "mysql":
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
	case "postgres":
		var postgresErr *pgconn.PgError
		if !errors.As(err, &postgresErr) {
			t.Fatalf("error = %T %v, want PgError", err, err)
		}
		if postgresErr.Code != "23514" {
			t.Errorf("PostgreSQL error code = %q, want 23514", postgresErr.Code)
		}
		if postgresErr.ConstraintName != constraint {
			t.Errorf("PostgreSQL constraint = %q, want %q", postgresErr.ConstraintName, constraint)
		}
	}
}
