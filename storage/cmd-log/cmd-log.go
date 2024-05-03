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
	sqlstatement := `INSERT INTO log(script, status, output, start, finish) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := s.db.QueryRowContext(ctx, sqlstatement, c.FullScript, c.CMDStatus, c.Output, c.Start, c.End).Scan(&c.ID)
	if err != nil {
		return fmt.Errorf("problems with saving command to log: %v", err)
	}
	return nil
}

func (s *CMDlog) Update(ctx context.Context, c *storage.Command) error {
	sqlstatement := `UPDATE log SET status = $1, output = $2, finish = $3 WHERE id = $4`
	_, err := s.db.ExecContext(ctx, sqlstatement, c.CMDStatus, c.Output, c.End, c.ID)
	if err != nil {
		return fmt.Errorf("can't update cmd: %v", err)
	}
	return nil
}
func (s *CMDlog) PickByID(ctx context.Context, id int) (*storage.Command, error) {
	sqlstatement := `SELECT * FROM log WHERE id = $1`
	var cmd storage.Command
	err := s.db.QueryRowContext(ctx, sqlstatement, id).Scan(&cmd.ID, &cmd.FullScript, &cmd.CMDStatus, &cmd.Output, &cmd.Start, &cmd.End)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("can't pick command with this id, maybe id is wrong")
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick cmd from log: %v", err)
	}
	return &cmd, nil
}
func (s *CMDlog) ShowAll(ctx context.Context) (*[]storage.Command, error) {
	sqlstatement := `SELECT id, script, status, start, finish FROM log`
	var commands []storage.Command
	rows, err := s.db.QueryContext(ctx, sqlstatement)

	if err != nil {
		return nil, fmt.Errorf("Problems with executing a query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cmd storage.Command
		err = rows.Scan(&cmd.ID, &cmd.FullScript, &cmd.CMDStatus, &cmd.Start, &cmd.End)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nothing in log")
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to scan a row from log: %v", err)
		}
		commands = append(commands, cmd)
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("No comamnds in log")
	}
	return &commands, nil
}
func (s *CMDlog) Close(ctx context.Context) error {
	sqlstatement := `DROP TABLE log`
	_, err := s.db.ExecContext(ctx, sqlstatement)
	if err != nil {
		return fmt.Errorf("can't truncate log while shutdown: %v", err)
	}
	return nil
}
func (s *CMDlog) Init() error {
	sqlstatement := `CREATE TABLE IF NOT EXISTS log (id SERIAL PRIMARY KEY, script TEXT, status TEXT,
		 output TEXT, start TIMESTAMP WITHOUT TIME ZONE, finish TIMESTAMP WITHOUT TIME ZONE)`
	_, err := s.db.Exec(sqlstatement)
	if err != nil {
		return fmt.Errorf("can't create a table: %v", err)
	}
	return nil
}
