package server

import (
	"context"

	"connectrpc.com/connect"

	"github.com/jamesread/armature-iam/password"
	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/store"
)

// Reserved API key name.
const bearerTokenKeyName = "Bearer token"

func (s *Server) GetBearerToken(ctx context.Context, _ *connect.Request[apiv1.GetBearerTokenRequest]) (*connect.Response[apiv1.GetBearerTokenResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.findBearerToken(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("get bearer token")
		return nil, mapStoreError(err)
	}
	if row != nil {
		return connect.NewResponse(&apiv1.GetBearerTokenResponse{Token: row.KeyValue}), nil
	}
	token, err := s.createBearerToken(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("create bearer token")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.GetBearerTokenResponse{Token: token}), nil
}

func (s *Server) RegenerateBearerToken(ctx context.Context, _ *connect.Request[apiv1.RegenerateBearerTokenRequest]) (*connect.Response[apiv1.RegenerateBearerTokenResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.findBearerToken(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("find bearer token")
		return nil, mapStoreError(err)
	}
	// Revokes the compromised token.
	if row != nil {
		if err := s.store.DeleteAPIKey(ctx, row.ID, au.User.ID); err != nil {
			s.log.WithError(err).Error("delete bearer token")
			return nil, mapStoreError(err)
		}
	}
	token, err := s.createBearerToken(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("regenerate bearer token")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.RegenerateBearerTokenResponse{Token: token}), nil
}

func (s *Server) findBearerToken(ctx context.Context, userID int) (*store.APIKeyRow, error) {
	rows, err := s.store.ListAPIKeysForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].Name == bearerTokenKeyName {
			return &rows[i], nil
		}
	}
	return nil, nil
}

func (s *Server) createBearerToken(ctx context.Context, userID int) (string, error) {
	secret, err := password.GenerateAPIKey("sa_")
	if err != nil {
		return "", err
	}
	// Read-only: displays never mutate.
	if _, err := s.store.CreateAPIKey(ctx, userID, bearerTokenKeyName, secret, true); err != nil {
		return "", err
	}
	return secret, nil
}
