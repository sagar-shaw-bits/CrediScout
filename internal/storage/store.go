package storage

import (
	"context"

	"github.com/sagar-shaw-bits/CrediScout/internal/domain"
)

type Store interface {
	Save(ctx context.Context, doc domain.Document) error
}
