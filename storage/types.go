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
	Save(ctx context.Context, c *Command) error
	PickByID(ctx context.Context, id int) (*Command, error)
	ShowAll(ctx context.Context) (*[]Command, error)
	Update(ctx context.Context, c *Command) error
	Close(ctx context.Context) error
}
type Command struct {
	ID         int       `json:"id"`
	FullScript string    `json:"script"`
	CMDStatus  string    `json:"cmdstatus"`
	Output     string    `json:"output"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
}
