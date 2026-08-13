package utils

import (
	"errors"
	"unicode"
)

// ValidatePasswordStrength 统一密码策略：6-72 位，且必须同时包含字母和数字。
// 会员注册/改密、管理员初始密码、管理员改密、后台重置密码共用同一规则。
func ValidatePasswordStrength(password string) error {
	if l := len(password); l < 6 || l > 72 {
		return errors.New("密码长度需在 6-72 位之间")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码必须同时包含字母和数字")
	}
	return nil
}
