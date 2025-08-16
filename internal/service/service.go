package service

import (
	"noteApp/internal/config"
	"noteApp/pkg/logger"
)

type HasherI interface {
	GenerateHash(password string) (string, error)
	ComparePassword(hash string, password string) error
}

type RepositoryI interface {
	AuthRI
	NoteRI
	UserRI
}

type Service struct {
	*AuthS
	*NoteS
	*UserS
}

func NewService(
	repos RepositoryI,
	hasher HasherI,
	cfg config.AuthCfg,
	log *logger.Logger,
) Service {
	return Service{
		AuthS: NewAuthService(repos, hasher, cfg, log),
		NoteS: NewNoteService(repos, log),
		UserS: NewUserService(repos, repos, hasher, log),
	}
}
