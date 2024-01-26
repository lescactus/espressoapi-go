package sql

import "time"

type Roaster struct {
	Id        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}
