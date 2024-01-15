package mysql

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
)

type MysqlDB struct {
	db *sqlx.DB
}

var _ repository.SheetRepository = (*MysqlDB)(nil)
var _ repository.BeansRepository = (*MysqlDB)(nil)

func New(db *sqlx.DB) *MysqlDB {
	return &MysqlDB{
		db: db,
	}
}

func (db *MysqlDB) CreateSheet(ctx context.Context, sheet *sql.Sheet) error {
	query := `INSERT INTO sheets (name) VALUES (?)`
	_, err := db.db.ExecContext(ctx, query, sheet.Name)
	if err != nil {
		// Checking if the entry inserted is a duplicate:
		// ERROR 1062 (23000): Duplicate entry 'xxxxx' for key 'yyyy'
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
			return repository.ErrSheetAlreadyExists
		}
		return fmt.Errorf("failed to insert record to the database: %w", err)
	}

	return nil
}

func (db *MysqlDB) GetSheetById(ctx context.Context, id int) (*sql.Sheet, error) {
	var s sql.Sheet

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE id = ?", id).StructScan(&s)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, repository.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet id=%d from the database: %w", id, err)
	}

	return &s, nil
}

func (db *MysqlDB) GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error) {
	var s sql.Sheet

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE name = ?", name).StructScan(&s)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, repository.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet name=\"%s\" from the database: %w", name, err)
	}

	return &s, nil
}

func (db *MysqlDB) GetAllSheets(ctx context.Context) ([]sql.Sheet, error) {
	var sheets = make([]sql.Sheet, 0)
	err := db.db.SelectContext(ctx, &sheets, "SELECT * FROM sheets")
	if err != nil {
		return sheets, fmt.Errorf("failed to read records for sheets: %w", err)
	}

	return sheets, nil
}

func (db *MysqlDB) UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error) {
	now := time.Now()
	sheet.Id = id
	sheet.UpdatedAt = &now

	// CreatedAt should be immutable
	res, err := db.db.ExecContext(ctx, `UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?`, sheet.Name, sheet.UpdatedAt, sheet.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record for sheet id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return nil, repository.ErrSheetDoesNotExist
	}

	return sheet, nil
}

func (db *MysqlDB) DeleteSheetById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM sheets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete record for sheet id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return repository.ErrSheetDoesNotExist
	}

	return nil
}

func (db *MysqlDB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

func (db *MysqlDB) CreateBeans(ctx context.Context, beans *sql.Beans) error {
	query := `INSERT INTO beans (roaster_name, beans_name, roast_date, roast_level) VALUES (?, ?, ?, ?)`
	_, err := db.db.ExecContext(ctx, query,
		beans.RoasterName, beans.BeansName, beans.RoastDate, beans.RoastLevel)
	if err != nil {
		return fmt.Errorf("failed to insert record to the database: %w", err)
	}

	return nil
}

func (db *MysqlDB) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var b sql.Beans

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM beans WHERE id = ?", id).StructScan(&b)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, repository.ErrBeansDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for beans id=%d from the database: %w", id, err)
	}

	return &b, nil
}

func (db *MysqlDB) GetAllBeans(ctx context.Context) ([]sql.Beans, error) {
	var beans = make([]sql.Beans, 0)
	err := db.db.SelectContext(ctx, &beans, "SELECT * FROM beans")
	if err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}

	return beans, nil
}

func (db *MysqlDB) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	beans.Id = id

	res, err := db.db.ExecContext(ctx, `UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?`,
		beans.RoasterName, beans.BeansName, beans.RoastDate, beans.RoastLevel, beans.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record for beans id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return nil, repository.ErrBeansDoesNotExist
	}

	return beans, nil
}

func (db *MysqlDB) DeleteBeansById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM beans WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete record for beans id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return repository.ErrBeansDoesNotExist
	}

	return nil
}
