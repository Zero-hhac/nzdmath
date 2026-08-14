package service

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"

	"math-top/internal/config"
)

// sendMail 通用邮件发送底层方法，支持 465 (SSL/TLS 直连) 及 25/587 (STARTTLS/明文) 端口。
func sendMail(to, subject, body string) error {
	cfg := config.GlobalConfig.SMTP
	if cfg.Host == "" || cfg.Port == 0 || cfg.User == "" {
		return fmt.Errorf("smtp 未配置")
	}

	encodedSubject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	msg := "From: " + cfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	if cfg.Port == 465 {
		// 465 端口采用 SSL/TLS 直连
		tlsConfig := &tls.Config{
			ServerName: cfg.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtp tls 连接失败: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("smtp 创建客户端失败: %w", err)
		}
		defer client.Quit()

		if auth != nil {
			if ok, _ := client.Extension("AUTH"); ok {
				if err = client.Auth(auth); err != nil {
					return fmt.Errorf("smtp 认证失败: %w", err)
				}
			}
		}

		if err = client.Mail(cfg.From); err != nil {
			return fmt.Errorf("smtp 设置发件人失败: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp 设置收件人失败: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp 写入数据失败: %w", err)
		}
		if _, err = w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("smtp 发送邮件正文失败: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("smtp 关闭数据通道失败: %w", err)
		}
	} else {
		// 25 / 587 等端口走默认 STARTTLS 模式
		if err := smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg)); err != nil {
			return fmt.Errorf("smtp 发送失败: %w", err)
		}
	}

	return nil
}

// SendPasswordResetCode 发送找回密码验证码邮件。
func SendPasswordResetCode(to, username, code string) error {
	subject := "【数学协会】密码找回验证码"
	body := fmt.Sprintf("你好 %s：\n\n你的密码找回验证码是：%s（10 分钟内有效）。\n如果这不是你本人的操作，请忽略本邮件。\n\n—— 数学协会", username, code)
	if err := sendMail(to, subject, body); err != nil {
		return err
	}
	slog.Info("找回密码验证码邮件已发送", "to", to)
	return nil
}

// SendRegisterVerifyCode 发送会员注册验证码邮件。
func SendRegisterVerifyCode(to, code string) error {
	subject := "【数学协会】会员注册验证码"
	body := fmt.Sprintf("你好：\n\n欢迎加入数学协会！你的会员注册验证码是：%s（10 分钟内有效）。\n如果这不是你本人的操作，请忽略本邮件。\n\n—— 数学协会", code)
	if err := sendMail(to, subject, body); err != nil {
		return err
	}
	slog.Info("会员注册验证码邮件已发送", "to", to)
	return nil
}
