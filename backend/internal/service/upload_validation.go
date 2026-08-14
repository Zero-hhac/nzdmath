package service

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

// 公共上传校验模块：扩展名白名单 + 魔数嗅探，供聊天室、私聊、资源上传共用。
// 任何上传入口都必须先过白名单，再用 validateUploadContent 校验内容与扩展名一致，
// 防止伪装 HTML / SVG 等可执行内容落盘后以 text/html 同源执行。

var chatImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var chatFileExts = map[string]bool{
	".txt": true, ".md": true, ".json": true, ".csv": true,
	".pdf": true, ".doc": true, ".docx": true, ".ppt": true, ".pptx": true,
}

var chatVideoExts = map[string]bool{
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true, ".flv": true, ".wmv": true,
}

// allowedChatExt 判断扩展名是否属于聊天/私聊允许的类型（图片 + 文档，视频一律拒绝）。
func allowedChatExt(ext string) bool {
	return chatImageExts[ext] || chatFileExts[ext]
}

// validateUploadContent 校验文件内容与扩展名一致：读取文件头 512 字节，
// 用 http.DetectContentType 嗅探真实类型，防止伪装文件上传。
func validateUploadContent(ext string, src io.Reader) error {
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return errors.New("读取文件失败")
	}
	ctype := http.DetectContentType(head[:n])

	if strings.HasPrefix(ctype, "text/html") {
		return errors.New("文件内容与扩展名不符")
	}
	if chatImageExts[ext] && !strings.HasPrefix(ctype, "image/") {
		return errors.New("文件内容与扩展名不符")
	}
	if ext == ".pdf" && ctype != "application/pdf" && ctype != "application/octet-stream" {
		return errors.New("文件内容与扩展名不符")
	}
	return nil
}
