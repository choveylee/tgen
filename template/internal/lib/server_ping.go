package lib

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
	"github.com/choveylee/thttp"
	"github.com/choveylee/tlog"
)

var (
	serverPingReqUrl string
	serverPingCount  int
)

// initServerPing validates startup probe settings before any service I/O.
func initServerPing(ctx context.Context) *terror.Terror {
	pingCount := tcfg.DefaultInt(tcfg.LocalKey("SERVER_PING_COUNT"), 3)
	if pingCount <= 0 {
		errMsg := tlog.E(ctx).Msg("Init server ping err (server ping count must be positive)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server ping count"), errMsg)

		return errx
	}

	serverPingHost := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("SERVER_PING_HOST"), "http://127.0.0.1"))
	if serverPingHost == "" {
		// An empty host in existing generated configs means the local service.
		serverPingHost = "http://127.0.0.1"
	} else if !strings.Contains(serverPingHost, "://") {
		serverPingHost = "http://" + serverPingHost
	}

	parsedUrl, err := url.Parse(serverPingHost)
	if err != nil || parsedUrl.Hostname() == "" || parsedUrl.User != nil ||
		(parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https") {
		errMsg := tlog.E(ctx).Msg("Init server ping err (server ping host must be an HTTP or HTTPS URL without credentials)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server ping host"), errMsg)

		return errx
	}

	host := parsedUrl.Hostname()
	hostIP := net.ParseIP(host)
	if host == "localhost" || (hostIP != nil && hostIP.IsLoopback()) {
		httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 8080)
		parsedUrl.Host = net.JoinHostPort(host, fmt.Sprint(httpPort))
	}

	serverPingReqUrl = parsedUrl.JoinPath("healthz").String()
	serverPingCount = pingCount

	return nil
}

// PingServer checks startup health using bounded retries for transient failures.
// The caller supplies the overall probe deadline, including waits between attempts.
func PingServer(ctx context.Context) *terror.Terror {
	retryOption := thttp.NewRetryTransOption().WithMaxCount(serverPingCount).WithWaitTime(time.Second, time.Second)
	requestOption := thttp.NewRequestOption().WithRetryTransOption(retryOption)
	response, err := thttp.Get(ctx, serverPingReqUrl, requestOption, nil)
	if err != nil {
		// net/http errors include the URL, whose query may contain credentials.
		var requestError *url.Error
		if errors.As(err, &requestError) {
			err = requestError.Err
		}

		errMsg := tlog.E(ctx).Err(err).Msgf("Ping server (ping count: %d) err (get %v)",
			serverPingCount, err)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errMsg := tlog.E(ctx).Msgf("Ping server (ping count: %d, status code: %d) err (health response must be HTTP 200)",
			serverPingCount, response.StatusCode)

		errx := terror.NewRawTerror(ctx, terror.ErrSvcExecute("http server"), errMsg)

		return errx
	}

	return nil
}
