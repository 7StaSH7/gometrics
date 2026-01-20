package grpcserver

import (
	"context"
	"net"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const realIPHeader = "x-real-ip"

// TrustedSubnetInterceptor validates the client IP against the trusted subnet.
func TrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	if trustedSubnet == "" {
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Log.Error("invalid trusted subnet", zap.String("subnet", trustedSubnet), zap.Error(err))
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		clientIP := extractClientIP(ctx)
		if clientIP == "" {
			logger.Log.Warn("missing client ip", zap.String("subnet", trustedSubnet))
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		ip := net.ParseIP(clientIP)
		if ip == nil || !ipNet.Contains(ip) {
			logger.Log.Warn("IP not in trusted subnet", zap.String("ip", clientIP), zap.String("subnet", trustedSubnet))
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		return handler(ctx, req)
	}
}

func extractClientIP(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(realIPHeader); len(values) > 0 {
			return values[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			return host
		}
		return p.Addr.String()
	}

	return ""
}
