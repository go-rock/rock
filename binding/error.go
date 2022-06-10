package binding

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni   *ut.UniversalTranslator
	Trans ut.Translator
)

type CommonError struct {
	Errors map[string]interface{} `json:"errors"`
}

func ValidatorError(err error) CommonError {
	res := CommonError{}
	res.Errors = make(map[string]interface{})
	if err != nil {
		switch errs := err.(type) {
		case validator.ValidationErrors:
			for _, e := range errs {
				transtr := e.Translate(Trans)
				f := strings.ToLower(e.StructField())
				res.Errors[f] = transtr
			}
		default:
			res.Errors["error"] = err.Error()
		}
	}
	return res
}

func InitBinding() {
	zhs := zh.New()
	uni = ut.New(zhs)

	Trans, _ = uni.GetTranslator("zh")

	if v, ok := Validator.Engine().(*validator.Validate); ok {
		if err := zh_translations.RegisterDefaultTranslations(v, Trans); err != nil {
			panic(err)
		}

		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return fld.Tag.Get("comment")
		})
	}
}
