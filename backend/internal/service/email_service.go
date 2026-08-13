package service

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"

	"math-top/internal/config"
)

// SendPasswordResetCode 发送找回密码验证码邮件。
// SMTP 未配置时返回错误（由调用方决定 debug 模式回显验证码）。
func SendPasswordResetCode(to, username, code string) error {
	cfg := config.GlobalConfig.SMTP
	if cfg.Host == "" || cfg.Port == 0 || cfg.User == "" {
		return fmt.Errorf("smtp 未配置")
	}
	subject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte("【数学协会】密码找回验证码")) + "?="
	body := fmt.Sprintf("你好 %s：\n\n你的密码找回验证码是：%s（10 分钟内有效）。\n如果这不是你本人的操作，请忽略本邮件。\n\n—— 数学协会", username, code)
	msg := "From: " + cfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	if err := smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp 发送失败: %w", err)
	}
	slog.Info("找回密码验证码邮件已发送", "to", to)
	return nil
}
