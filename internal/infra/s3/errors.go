package s3

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func isNotFoundError(err error, target *types.NotFound) bool {
	return errors.As(err, &target)
}
