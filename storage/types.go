package storage

import (
	"context"
	"time"
)

type State int

const (
	Success State = iota
	Running
	Failed
)

func (s State) String() string {
	return [...]string{"Success", "Running", "Failed"}[s]
}

type Storage interface {
	Save(ctx context.Context, c *Command)
	PickByID(ctx context.Context, id int)
	ShowAll(ctx context.Context) (*[]Command, error)
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, c *Command) error
}
type Command struct {
	ID         int       `json:"id"`
	FullScript string    `json:"script"`
	CMDStatus  string    `json:"cmdstatus"`
	Output     string    `json:"output"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
}
