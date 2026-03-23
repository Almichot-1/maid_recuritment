package domain

import "context"

type MaidRepository interface {
	List(ctx context.Context) ([]Maid, error)
}
