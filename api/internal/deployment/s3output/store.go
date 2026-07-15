package s3output

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/friendsofshopware/shopmon/api/internal/storage"
)

type Store struct {
	storage *storage.S3Storage
}

var _ deployment.OutputStore = (*Store)(nil)

func NewStore(storage *storage.S3Storage) *Store {
	return &Store{storage: storage}
}

func (s *Store) GetDeploymentOutput(ctx context.Context, deploymentID int) (string, error) {
	output, err := s.storage.GetDeploymentOutput(ctx, deploymentID)
	if isNotFound(err) {
		return "", deployment.ErrOutputNotFound
	}
	return output, err
}

func (s *Store) DeleteDeploymentOutput(ctx context.Context, deploymentID int) error {
	return s.storage.DeleteDeploymentOutput(ctx, deploymentID)
}

func (s *Store) PresignUpload(ctx context.Context, deploymentID int) (string, error) {
	return s.storage.PresignUpload(ctx, deploymentID)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		switch apiError.ErrorCode() {
		case "NoSuchKey", "NotFound", "404":
			return true
		}
	}
	return false
}
