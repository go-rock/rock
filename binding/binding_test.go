package binding

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
)

type userReq struct {
	Name  string `binding:"required"`
	Email string `binding:"required,email"`
}

func TestMain(m *testing.M) {
	if err := InitBinding(); err != nil {
		fmt.Fprintln(os.Stderr, "init binding:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestValidateValid(t *testing.T) {
	err := Validate(&userReq{Name: "x", Email: "a@b.com"})
	if err != nil {
		t.Errorf("合法结构不应报错, got %v", err)
	}
}

func TestValidateInvalid(t *testing.T) {
	err := Validate(&userReq{})
	if err == nil {
		t.Fatal("空结构应校验失败")
	}
	if _, ok := err.(validator.ValidationErrors); !ok {
		t.Errorf("错误类型应为 validator.ValidationErrors, got %T", err)
	}
}

func TestValidateNonStruct(t *testing.T) {
	// 非结构体/指针结构体应跳过校验
	if err := Validate("not a struct"); err != nil {
		t.Errorf("字符串应跳过校验, got %v", err)
	}
	if err := Validate(42); err != nil {
		t.Errorf("数字应跳过校验, got %v", err)
	}
	if err := Validate([]string{"a"}); err != nil {
		t.Errorf("切片应跳过校验, got %v", err)
	}
	if err := Validate(nil); err != nil {
		t.Errorf("nil 应跳过校验, got %v", err)
	}
}

func TestDefaultValidatorEngine(t *testing.T) {
	if _, ok := Validator.Engine().(*validator.Validate); !ok {
		t.Error("Engine() 应返回 *validator.Validate")
	}
}

func TestValidatorError(t *testing.T) {
	// 无错误 → 空 map
	ce := ValidatorError(nil)
	if len(ce.Errors) != 0 {
		t.Errorf("无错误时 Errors 应为空, got %v", ce.Errors)
	}

	// 有校验错误 → 生成翻译后的字段错误
	ce = ValidatorError(Validate(&userReq{}))
	if len(ce.Errors) == 0 {
		t.Error("校验错误应生成 Errors map")
	}
	if _, ok := ce.Errors["name"]; !ok {
		t.Errorf("应包含 name 字段错误, got %v", ce.Errors)
	}
	if _, ok := ce.Errors["email"]; !ok {
		t.Errorf("应包含 email 字段错误, got %v", ce.Errors)
	}
}

func TestValidatorErrorUnknownType(t *testing.T) {
	ce := ValidatorError(&dummyError{msg: "boom"})
	if ce.Errors["error"] != "boom" {
		t.Errorf("未知错误类型应进 error 键, got %v", ce.Errors)
	}
}

type dummyError struct{ msg string }

func (d *dummyError) Error() string { return d.msg }
