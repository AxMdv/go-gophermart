package gophermart

import (
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type GophermartService struct {
	repository *storage.DBRepository
}

func New(dbRepository *storage.DBRepository) *GophermartService {
	return &GophermartService{repository: dbRepository}
}
