package main

import (
	"context"
	"errors"
	"log/slog"
	"math-top/internal/router"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

func main() {
	r := router.NewEngine()
	srv := &http.Server{
		Addr:    router.Addr(),
		Handler: r,
	}

	go func() {
		slog.Info("服务启动", "addr", srv.Addr)
		for _, ip := range listLocalIPs() {
			slog.Info("局域网访问", "url", "http://"+ip+srv.Addr)
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("服务异常退出", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，开始关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("服务关闭超时", "err", err)
		os.Exit(1)
	}
	slog.Info("服务已退出")
}

func listLocalIPs() []string {
	ips := map[string]struct{}{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ips[ip.String()] = struct{}{}
		}
	}
	out := make([]string, 0, len(ips))
	for ip := range ips {
		out = append(out, ip)
	}
	sort.Strings(out)
	return out
}