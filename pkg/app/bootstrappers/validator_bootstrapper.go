package bootstrappers

import (
	"regexp"

	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type ValidatorBootstrapper struct{}

var namespaceRegex = regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?")

var namespaceValidator validator.Func = func(fl validator.FieldLevel) bool {
	ns, ok := fl.Field().Interface().(string)
	if ok {
		match := namespaceRegex.Match([]byte(ns))
		if match {
			return true
		}
	}
	return false
}

func (v *ValidatorBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("namespace", namespaceValidator); err != nil {
			return err
		}
	}

	return nil
}
