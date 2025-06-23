package H

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	validator "github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"

	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	es_translations "github.com/go-playground/validator/v10/translations/es"
)

type (
	FieldTranslate  map[string]string
	ModelTranslate  map[string]FieldTranslate
	CustomValidator struct {
		Uni        *ut.UniversalTranslator
		ListModels map[string]ModelTranslate
	}
	GenericError struct {
		Message string      `json:"message"`
		Error   interface{} `json:"details_error,omitempty"`
	}
	GenericMessage struct {
		Message string `json:"message"`
	}
	ValidateStruct struct {
		Context echo.Context
		Data    interface{}
	}
)

var validator_instance *validator.Validate
var mutex *sync.Mutex
var last_language string

func SnakeCase(s string) string {
	var snake string
	last_space := false
	s = strings.Replace(s, "UUID", "Uuid", -1)
	s = strings.Replace(s, "ID", "Id", -1)
	for i, r := range s {
		if i == 0 {
			snake += strings.ToLower(string(r))
		} else {
			if !last_space && 'A' <= r && r <= 'Z' {
				snake += "_" + strings.ToLower(string(r))
				last_space = true
			} else {
				snake += string(r)
			}
		}
		if 'a' <= r && r <= 'z' {
			last_space = false
		}
	}
	return snake
}
func (cv *CustomValidator) Validate(i interface{}) error {
	to_validate := i.(ValidateStruct)
	lang := "es"
	if h_lang := to_validate.Context.Request().Header.Get("X-Language"); h_lang != "" {
		switch h_lang {
		case "es", "en":
			lang = h_lang
		default:
			lang = "es"
		}
	}
	trans, _ := cv.Uni.GetTranslator(lang)
	fTranslation := make(map[string]string)
	typeOf := reflect.TypeOf(to_validate.Data).Elem()
	modelName := typeOf.Name()
	if modelTranslate, ok := cv.ListModels[modelName]; ok {
		if fieldTranslate, ok := modelTranslate[lang]; ok {
			fTranslation = fieldTranslate
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	validator_instance.RegisterTagNameFunc(func(field reflect.StructField) string {
		if name, ok := fTranslation[field.Name]; ok {
			return name
		}
		return field.Name
	})
	if err := validator_instance.Struct(to_validate.Data); err != nil {
		var list_error []map[string]interface{}
		if last_language != lang {
			if lang == "es" {
				es_translations.RegisterDefaultTranslations(validator_instance, trans)
			} else {
				en_translations.RegisterDefaultTranslations(validator_instance, trans)
			}
			last_language = lang
		}
		for _, err := range err.(validator.ValidationErrors) {
			el := make(map[string]interface{})
			el["content"] = err.Value()
			el["rule"] = err.Tag()
			el["field"] = SnakeCase(err.StructField())
			el["field_lang"] = err.Field()
			el["message"] = err.Translate(trans)
			list_error = append(list_error, el)
		}

		return echo.NewHTTPError(http.StatusBadRequest, list_error)
	}
	return nil
}

func init() {
	validator_instance = validator.New()
	mutex = new(sync.Mutex)
	last_language = ""
}
