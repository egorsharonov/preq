package validators

import (
	"context"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

type DTOValidator interface {
	Validate(ctx context.Context) *model.APIError
}
