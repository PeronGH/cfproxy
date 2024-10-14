package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudflare/cloudflared/carrier"
	"github.com/rs/zerolog"
)

const (
	LogFieldHost               = "host"
	cfAccessClientIDHeader     = "Cf-Access-Client-Id"
	cfAccessClientSecretHeader = "Cf-Access-Client-Secret"
)

type connectOptions struct {
	target   string
	local    string
	logger   *zerolog.Logger
	clientId string
	clientSecret   string
}

func connect(options *connectOptions) error {
	targetURL, err := url.Parse(options.target)
	if err != nil {
		return err
	}
	if targetURL.Scheme == "" {
		targetURL.Scheme = "https"
	}

	wsConn := carrier.NewWSConnection(options.logger)

	forward := func(client io.ReadWriter, destination string) {
		headers := make(http.Header)
		if options.clientId != "" {
			headers.Set(cfAccessClientIDHeader, options.clientId)
		}

		if options.clientSecret != "" {
			headers.Set(cfAccessClientSecretHeader, options.clientSecret)
		}

		carrier.SetBastionDest(headers, destination)

		err = wsConn.ServeStream(&carrier.StartOptions{
			Host:      targetURL.Host,
			OriginURL: targetURL.String(),
			Headers:   headers,
		}, client)

		if err != nil {
			options.logger.Error().Err(err).Msg("Error in carrier stream")
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle non-tunneling requests
		if r.Method != http.MethodConnect {
			destination := r.Host
			if !strings.Contains(destination, ":") {
				destination += ":80"
			}

			// TODO: implement non-CONNECT proxy
			http.Error(w, "Non-CONNECT proxy not implemented", http.StatusNotImplemented)
			return
		}

		// Handle CONNECT requests
		w.WriteHeader(http.StatusOK)

		hijacker := w.(http.Hijacker)
		client, _, err := hijacker.Hijack()
		if err != nil {
			options.logger.Error().Err(err).Msg("Failed to hijack connection")
			return
		}
		defer client.Close()

		forward(client, r.Host)
	})

	return http.ListenAndServe(options.local, handler)
}
