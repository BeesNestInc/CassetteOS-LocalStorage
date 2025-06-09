package v2

import (
	"github.com/BeesNestInc/CassetteOS-LocalStorage/codegen"
)

type LocalStorage struct{}

func NewLocalStorage() codegen.ServerInterface {
	return &LocalStorage{}
}
