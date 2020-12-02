package Week02

import (
	"database/sql"
	"errors"
	"fmt"

	pkgError "github.com/pkg/errors"
)

/*
我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？
*/

type XModel struct{}

var mockQueryXModel = func() (*XModel, error) {
	return nil, sql.ErrNoRows
}

type XModelDao struct{}

func (*XModelDao) getXModel(id int) (*XModel, error) {
	// mock query XModel and get sql.ErrNoRows
	xModel, err := mockQueryXModel()
	if err != nil {
		// errors from standard package should be wrapped with extra messages.
		return nil, pkgError.Wrap(err, fmt.Sprintf("get XModel(%v) failed", id))
	}
	return xModel, nil
}

var ErrServiceTemporaryUnavailable = errors.New("service temporary unavailable")

type XModelService struct{}

func (*XModelService) getXModel(id int) (*XModel, error) {
	xModelDao := &XModelDao{}
	xModel, err := xModelDao.getXModel(id)
	if err == nil {
		return xModel, nil
	}
	// we should return error if no data is an error for user from business point of view.
	// but we don't think no data is an error for user at present, so we just tell user there is no data.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	// log error and return service temporary unavailable for other erorrs.
	fmt.Printf("%+v\n", err)
	// sentinel error should be wrapped with fmt.Errorf("%w")
	return nil, fmt.Errorf("%w", ErrServiceTemporaryUnavailable)
}
