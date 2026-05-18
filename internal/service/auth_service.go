package service

import (
	"context"

	"drone-management/internal/domain"
	"drone-management/internal/repo"
	"drone-management/internal/utils"
)

type AuthService struct {
	principals *repo.PrincipalRepo
	drones     *repo.DroneRepo
	signer     *utils.JWTSigner
	clock      utils.Clock
}

func NewAuthService(p *repo.PrincipalRepo, d *repo.DroneRepo, signer *utils.JWTSigner, clock utils.Clock) *AuthService {
	return &AuthService{principals: p, drones: d, signer: signer, clock: clock}
}

type IssueTokenResult struct {
	Token     string
	Principal *domain.Principal
}

func (s *AuthService) IssueToken(ctx context.Context, name string, role domain.Role) (*IssueTokenResult, error) {
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	if !role.Valid() {
		return nil, domain.ErrInvalidInput
	}
	p, err := s.principals.Upsert(ctx, name, role)
	if err != nil {
		return nil, err
	}
	if role == domain.RoleDrone {
		if _, err := s.drones.EnsureForPrincipal(ctx, p.ID); err != nil {
			return nil, err
		}
	}
	tok, err := s.signer.Issue(p.ID, p.Name, p.Role, s.clock.Now())
	if err != nil {
		return nil, err
	}
	return &IssueTokenResult{Token: tok, Principal: p}, nil
}
