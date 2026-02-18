package factory

import (
	"github.com/complytime/gemara-content-service/mapper"
	"github.com/complytime/gemara-content-service/mapper/plugins/basic"
)

func MapperByID(_ mapper.ID) mapper.Mapper {
	return basic.NewBasicMapper()
}
