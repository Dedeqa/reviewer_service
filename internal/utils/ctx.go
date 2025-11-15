package utils

import (
	"context"
	"time"
)

func CtxTimeout(d time.Duration) (context.Context, func()) {
	return context.WithTimeout(context.Background(), d)
}
