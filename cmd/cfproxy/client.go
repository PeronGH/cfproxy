package main

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/cloudflare/cloudflared/carrier"
	"github.com/rs/zerolog"
)

const (
	cfAccessClientIDHeader     = "Cf-Access-Client-Id"
	cfAccessClientSecretHeader = "Cf-Access-Client-Secret"
)

type connectOptions struct {
	target       string
	local        string
	logger       *zerolog.Logger
	clientId     string
	clientSecret string
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
			// Modify request and dump it
			reqURL, err := url.Parse(r.RequestURI)
			if err != nil || reqURL.Scheme != "http" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			r.RequestURI = ""
			r.URL = reqURL

			destination := reqURL.Host
			if reqURL.Port() == "" {
				destination += ":80"
			}

			reqBytes, err := httputil.DumpRequest(r, false)
			if err != nil {
				http.Error(w, "Failed to read request", http.StatusInternalServerError)
				return
			}

			hijacker := w.(http.Hijacker)
			client, _, err := hijacker.Hijack()
			if err != nil {
				options.logger.Error().Err(err).Msg("Failed to hijack connection")
				return
			}
			defer client.Close()

			// create prepended reader
			pr := newPrependedReader(client, reqBytes)
			rw := newReaderWriter(pr, client)

			forward(rw, destination)
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
