package middleware

import (
	"context"
	"net/http"
)
const HeaderKeyXForwardedFor = "X-Forwarded-Authorization"

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
		
        if xForwardedFor := r.Header.Get(HeaderKeyXForwardedFor); xForwardedFor != "" {
            ctx = setXForwardedFor(ctx, xForwardedFor)
        }
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setXForwardedFor(ctx context.Context, xForwardedFor string) context.Context{
	updatedCtx := context.WithValue(ctx, "xForwardedFor", xForwardedFor)
	return updatedCtx
}

func setHeaders(ctx context.Context, request *http.Request) error {
    // if encodedInternalIDToken, err := middleware.GetEncodedInternalIDToken(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyZozoInternalIDToken, encodedInternalIDToken)
    // }
    // if ipAddress, err := middleware.GetUserIPAddress(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyUserIP, ipAddress)
    // }
    // if userAgent, err := middleware.GetForwardedUserAgent(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyForwardedUserAgent, userAgent)
    // }
    // if traceID, err := middleware.GetTraceID(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyZozoTraceID, traceID)
    // }
    // if uid, err := middleware.GetUID(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyZozoUID, uid)
    // }
    // if apiClient, err := middleware.GetAPIClient(ctx); err == nil {
    //     request.Header.Set(constant.HeaderKeyAPIClient, apiClient)
    // }
    // if xForwardedFor, err := middleware.GetXForwardedFor(ctx); err == nil {
    //     if remoteAddr, err := middleware.GetRemoteAddress(ctx); err == nil {
    //         host, _, e := net.SplitHostPort(remoteAddr)
    //         if e != nil {
    //             return xerrors.Errorf("split remote address: %v", e)
    //         }
    //         request.Header.Set(constant.HeaderKeyXForwardedFor, xForwardedFor+", "+host)
    //     }
    // }
    return nil
}