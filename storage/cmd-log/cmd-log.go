package cmdlog

import (
	"Term-api/storage"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type CMDlog struct {
	db *sql.DB
}

func New(psqlconnection string) (*CMDlog, error) {
	db, err := sql.Open("postgres", psqlconnection)
	if err != nil {
		return nil, fmt.Errorf("can't open DB: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("trouble with connection to DB: %v", err)
	}
	return &CMDlog{db: db}, nil
}
func (s *CMDlog) Save(ctx context.Context, c *storage.Command) error {
	sqlstatement := `INSERT INTO log(script, success, output, createdAt) VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, sqlstatement, c.FullScript, c.Success, c.Output, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("problems with saving command to log: %v", err)
	}
	return nil
}
func (s *CMDlog) PickByID(ctx context.Context, id int) (*storage.Command, error) {
	sqlstatement := `SELECT * FROM log WHERE id = $1`
	var cmd storage.Command
	err := s.db.QueryRowContext(ctx, sqlstatement, id).Scan(&cmd.ID, &cmd.FullScript, &cmd.Success, &cmd.Output, &cmd.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("can't pick command with this id, maybe id is wrong")
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick cmd from log: %v", err)
	}
	return &cmd, nil
}
func (s *CMDlog) ShowALL(ctx context.Context) (*[]storage.Command, error) {
	sqlstatement := `SELECT * FROM log`
	var commands []storage.Command
	rows, err := s.db.QueryContext(ctx, sqlstatement)

	if err != nil {
		return nil, fmt.Errorf("Problems with executing a query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cmd storage.Command
		err = rows.Scan(&cmd.ID, &cmd.FullScript, &cmd.Success, &cmd.Output, &cmd.CreatedAt)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nothing in log")
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to scan a row from log: %v", err)
		}
		commands = append(commands, cmd)
	}
	return &commands, nil
}
func (s *CMDlog) Init(ctx context.Context) error {
	sqlstatement := `CREATE TABLE IF NOT EXISTS log (id SERIAL, script TEXT, success BOOL, output TEXT, createdat TIME)`
	_, err := s.db.ExecContext(ctx, sqlstatement)
	if err != nil {
		return fmt.Errorf("can't create a table: %v", err)
	}
	return nil
}
