package cache

import (
	"fmt"
	"os"
)

type fsCache struct {
	Dir string
}

func (c *fsCache) Name() string {
	return c.Dir
}

func (c *fsCache) Clear() error {
	if err := os.RemoveAll(c.Dir); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	if err := os.Mkdir(c.Dir, 0777); err != nil {
		return fmt.Errorf("failed to make new one: %w", err)
	}

	return nil
}
