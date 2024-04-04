package storage

import (
	"bytes"
	"context"
	"log"
	"os/exec"
	"time"
)

type Storage interface {
	Save(ctx context.Context, c *Command)
	PickByID(ctx context.Context, id int)
	ShowALL(ctx context.Context) (*[]Command, error)
}
type Command struct {
	ID         int       `json:"id"`
	FullScript string    `json:"script"`
	Success    bool      `json:"success"`
	Output     string    `json:"output"`
	CreatedAt  time.Time `json:"time"`
	IsParallel bool      `json:"parallel"`
}

func (c *Command) ScriptToCmd() (*exec.Cmd, error) {
	cmd := exec.Command("bash")
	stdin := bytes.NewBufferString(c.FullScript)
	cmd.Stdin = stdin
	log.Println(c.FullScript)
	return cmd, nil
}
