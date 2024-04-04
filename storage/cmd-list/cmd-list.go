package cmdlist

import (
	"Term-api/storage"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type List struct {
	db *sql.DB
}

/*	Надо сделать отдельное поле, которое будет давать команде возможность выполняться конкурентно, а не подряд
	типо флажок parallel - true/false или как-то так */

func New(psqlconnection string) (*List, error) {
	db, err := sql.Open("postgres", psqlconnection)
	if err != nil {
		return nil, fmt.Errorf("can't open DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("connnection to DB failed: %v", err)
	}
	return &List{db: db}, nil
}

func (l *List) Save(ctx context.Context, c *storage.Command) error {
	sqlstatement := `INSERT INTO list(script, createdAt, isparallel) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := l.db.QueryRowContext(ctx, sqlstatement, c.FullScript, c.CreatedAt, c.IsParallel).Scan(&id)
	if err != nil {
		return fmt.Errorf("Troubles with saving command to list: %v", err)
	}
	c.ID = id
	return nil
}

func (l *List) PickByID(ctx context.Context, id int) (*storage.Command, error) {
	sqlstatement := `SELECT script, createdAt, isparallel FROM list WHERE id = $1`
	cmd := storage.Command{ID: id}
	err := l.db.QueryRowContext(ctx, sqlstatement, cmd.ID).Scan(&cmd.FullScript, &cmd.CreatedAt, &cmd.IsParallel)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("can't find command with this id from list")
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick cmd from list: %v", err)
	}
	return &cmd, nil
}

func (l *List) ShowALL(ctx context.Context) (*[]storage.Command, error) {
	sqlstatement := `SELECT * FROM list`
	var commands []storage.Command
	rows, err := l.db.QueryContext(ctx, sqlstatement)
	if err != nil {
		return nil, fmt.Errorf("trouble with scanning list: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cmd storage.Command
		err := rows.Scan(&cmd.ID, &cmd.FullScript, &cmd.CreatedAt, &cmd.IsParallel)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("No commands in list")
		}
		if err != nil {
			return nil, fmt.Errorf("can't scan command from list: %v", err)
		}
		commands = append(commands, cmd)
	}
	return &commands, nil
}
func (l *List) Init(ctx context.Context) error {
	sqlstatement := `CREATE TABLE IF NOT EXISTS list (id SERIAL, script TEXT, createdAT TIME, isparallel BOOLEAN)`
	_, err := l.db.ExecContext(ctx, sqlstatement)
	if err != nil {
		return fmt.Errorf("can't create table list: %v", err)
	}
	return nil
}
