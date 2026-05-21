package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
)

const (
	userSearchMinLen = 1
	userSearchLimit  = 10
)

type UserService interface {
	Search(ctx context.Context, callerID, query string) ([]*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// Search returns up to userSearchLimit users matching the trimmed query in
// name or email. The caller themselves are never returned so invite pickers
// don't suggest the project owner to themselves.
func (s *userService) Search(ctx context.Context, callerID, query string) ([]*model.User, error) {
	q := strings.TrimSpace(query)
	if len(q) < userSearchMinLen {
		return nil, fmt.Errorf("%w: q is required", ErrValidation)
	}
	return s.userRepo.Search(ctx, q, callerID, userSearchLimit)
}
