package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	appconfig "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	sftpbase "gitlab.services.mts.ru/salsa/go-base/application/infrastructure/sftp"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/datastructures"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

type IFTPService interface {
	sftpbase.ClientBase
	CheckDocumentAccess(ctx context.Context, relativePath string) *model.APIError
}

type FTPService struct {
	sftpbase.Client
}

func NewFTPService(cfg *appconfig.FTPConfig) (*FTPService, error) {
	baseClient, err := sftpbase.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &FTPService{
		*baseClient,
	}, nil
}

func (s *FTPService) CheckDocumentAccess(ctx context.Context, relativePath string) *model.APIError {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "FTPService.CheckDocumentAccess")
	defer span.End()

	log := diagnostics.LoggerFromContext(ctx)

	span.SetAttributes(attribute.String("relativePath", relativePath))

	if !s.validFileExtension(relativePath) {
		return model.ErrInvalidParameterValue("contract.relativePath должен указывать на файл с допустимым расширением")
	}

	if err := s.checkFTPFileExists(ctx, relativePath); err != nil {
		log.Error("FTP file check failed", zap.String("fullPath", relativePath), zap.Error(err))

		switch {
		case errors.Is(err, os.ErrPermission):
			return model.ErrDocNotAllowed
		case errors.Is(err, os.ErrNotExist):
			return model.ErrDocObjectNotFound
		default:
			return model.ErrDocObjectNotFound
		}
	}

	return nil
}

func (s *FTPService) validFileExtension(relativePath string) bool {
	ext := strings.ToLower(path.Ext(relativePath))
	allowedExtensions := datastructures.HashSet[string]{
		".pdf":  {},
		".doc":  {},
		".docx": {},
		".jpg":  {},
		".jpeg": {},
		".png":  {},
	}

	return allowedExtensions.Contains(ext)
}

func (s *FTPService) checkFTPFileExists(ctx context.Context, filePath string) error {
	client, cleanup, err := s.GetAndOpenSFTPConnection(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = client.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found or not accessible: %w", err)
	}

	return nil
}
