package service

import (
	"context"
	"dcs/internal/config"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repository.UserRepo
	authSrcRepo repository.UserAuthSourceRepo
	cfg         *config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepo, authSrcRepo repository.UserAuthSourceRepo, cfg *config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		authSrcRepo: authSrcRepo,
		cfg:         cfg,
	}
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	if username == "" || password == "" {
		return nil, errors.Wrap(nil, 40001, 400, "username and password are required")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to query user")
	}
	if user == nil {
		return nil, errors.Wrap(nil, 40101, 401, "invalid username or password")
	}
	if user.Status != "active" {
		return nil, errors.Wrap(nil, 40301, 403, "user account is disabled")
	}

	authSrc, err := s.authSrcRepo.GetByUserIDAndType(ctx, user.UserID, "local")
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to query auth source")
	}
	if authSrc == nil {
		return nil, errors.Wrap(nil, 40101, 401, "invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(authSrc.PasswordHash), []byte(password)); err != nil {
		return nil, errors.Wrap(nil, 40101, 401, "invalid username or password")
	}

	token, refreshToken, expiresIn, err := s.generateTokenPair(user.UserID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.UserID)

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*LoginResponse, error) {
	if refreshTokenStr == "" {
		return nil, errors.Wrap(nil, 40001, 400, "refresh_token is required")
	}

	claims, err := s.parseToken(refreshTokenStr)
	if err != nil {
		return nil, errors.Wrap(nil, 40101, 401, "invalid or expired refresh token")
	}
	if claims.TokenType != "refresh" {
		return nil, errors.Wrap(nil, 40101, 401, "invalid token type")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to query user")
	}
	if user == nil {
		return nil, errors.Wrap(nil, 40101, 401, "user not found")
	}
	if user.Status != "active" {
		return nil, errors.Wrap(nil, 40301, 403, "user account is disabled")
	}

	token, refreshToken, expiresIn, err := s.generateTokenPair(user.UserID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) generateTokenPair(userID, username, role string) (string, string, int, error) {
	expireHours := s.cfg.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}
	refreshDays := s.cfg.RefreshDays
	if refreshDays <= 0 {
		refreshDays = 7
	}

	now := time.Now()

	accessClaims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.cfg.Issuer,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	tokenStr, err := accessToken.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", "", 0, errors.Wrap(err, 50001, 500, "failed to sign access token")
	}

	refreshClaims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(refreshDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.cfg.Issuer,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", "", 0, errors.Wrap(err, 50001, 500, "failed to sign refresh token")
	}

	return tokenStr, refreshTokenStr, expireHours * 3600, nil
}

func (s *AuthService) parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(nil, 40101, 401, "unexpected signing method")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, 40101, 401, "invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.Wrap(nil, 40101, 401, "invalid token claims")
	}
	return claims, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, errors.Wrap(nil, 40101, 401, "invalid token type")
	}
	return claims, nil
}
