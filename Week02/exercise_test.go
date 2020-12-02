package Week02

import (
	"errors"
	"testing"

	pkgError "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestXModelService(t *testing.T) {
	// no data
	{
		xModelService := &XModelService{}
		xModel, err := xModelService.getXModel(0)
		assert.Nil(t, err)
		assert.Nil(t, xModel)
	}

	// one data
	{
		var expectModel = &XModel{}
		mockQueryXModel = func() (*XModel, error) {
			return expectModel, nil
		}

		xModelService := &XModelService{}
		xModel, err := xModelService.getXModel(0)
		assert.Nil(t, err)
		assert.Equal(t, expectModel, xModel)
	}

	// query error
	{
		var mockErr = pkgError.New("mock query error")
		mockQueryXModel = func() (*XModel, error) {
			return nil, mockErr
		}

		xModelService := &XModelService{}
		xModel, err := xModelService.getXModel(0)
		assert.Nil(t, xModel)
		assert.NotNil(t, err)
		assert.NotEqual(t, mockErr, err)
		assert.True(t, errors.Is(err, ErrServiceTemporaryUnavailable))
	}
}
