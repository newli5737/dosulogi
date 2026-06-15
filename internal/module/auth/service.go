package auth

import (
	"context"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/google/uuid"
)

type Service struct {
	repo   *Repository
	jwtMgr *util.JWTManager
}

func NewService(repo *Repository, jwtMgr *util.JWTManager) *Service {
	return &Service{repo: repo, jwtMgr: jwtMgr}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, string, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", ErrInvalidCred
	}
	if !u.IsActive {
		return nil, "", ErrInactive
	}
	if !util.VerifyPassword(u.Password, req.Password) {
		return nil, "", ErrInvalidCred
	}

	access, err := s.jwtMgr.SignAccess(u.ID.String(), u.Role)
	if err != nil {
		return nil, "", err
	}
	refresh, err := s.jwtMgr.SignRefresh(u.ID.String())
	if err != nil {
		return nil, "", err
	}

	hash := util.HashToken(refresh)
	expires := time.Now().Add(s.jwtMgr.RefreshTTL())
	if err := s.repo.SaveRefreshToken(ctx, u.ID, hash, expires); err != nil {
		return nil, "", err
	}

	return &LoginResponse{
		AccessToken: access,
		User: UserBrief{
			ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role,
		},
	}, refresh, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginResponse, string, error) {
	claims, err := s.jwtMgr.ParseRefresh(refreshToken)
	if err != nil {
		return nil, "", err
	}
	hash := util.HashToken(refreshToken)
	rt, err := s.repo.FindRefreshToken(ctx, hash)
	if err != nil {
		return nil, "", err
	}

	userID, _ := uuid.Parse(claims.UserID)
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil || !u.IsActive {
		return nil, "", ErrInvalidCred
	}
	_ = rt

	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, "", err
	}

	access, err := s.jwtMgr.SignAccess(u.ID.String(), u.Role)
	if err != nil {
		return nil, "", err
	}
	newRefresh, err := s.jwtMgr.SignRefresh(u.ID.String())
	if err != nil {
		return nil, "", err
	}
	newHash := util.HashToken(newRefresh)
	expires := time.Now().Add(s.jwtMgr.RefreshTTL())
	if err := s.repo.SaveRefreshToken(ctx, u.ID, newHash, expires); err != nil {
		return nil, "", err
	}

	return &LoginResponse{
		AccessToken: access,
		User: UserBrief{
			ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role,
		},
	}, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repo.RevokeRefreshToken(ctx, util.HashToken(refreshToken))
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !util.VerifyPassword(u.Password, req.OldPassword) {
		return ErrInvalidCred
	}
	hash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, userID, hash)
}

func (s *Service) ListUsers(ctx context.Context, role string, limit, offset int) ([]User, int, error) {
	return s.repo.List(ctx, role, limit, offset)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := &User{Email: req.Email, Password: hash, FullName: req.FullName, Role: req.Role, IsActive: true}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.FullName != nil {
		u.FullName = *req.FullName
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
