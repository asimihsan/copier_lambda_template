package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

// slackVerification holds the injected settings.
type slackVerification struct {
	signingSecret string
	logger        zerolog.Logger
}

// NewSlackVerificationMiddleware returns a middleware function that verifies Slack signatures.
func NewSlackVerificationMiddleware(signingSecret string, logger zerolog.Logger) func(next http.Handler) http.Handler {
	mw := &slackVerification{
		signingSecret: signingSecret,
		logger:        logger,
	}
	return mw.middleware
}

func (mw *slackVerification) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			mw.logger.Error().Err(err).Msg("Failed to read request body")
			http.Error(w, "Failed to read request", http.StatusInternalServerError)
			return
		}
		// Reset the body so the next handler can read it.
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		verifier, err := slack.NewSecretsVerifier(r.Header, mw.signingSecret)
		if err != nil {
			mw.logger.Error().Err(err).Msg("Failed to create secrets verifier")
			http.Error(w, "Failed to create secrets verifier", http.StatusInternalServerError)
			return
		}
		if _, err := verifier.Write(bodyBytes); err != nil {
			mw.logger.Error().Err(err).Msg("Failed to write body to verifier")
			http.Error(w, "Failed to write body to verifier", http.StatusInternalServerError)
			return
		}
		if err = verifier.Ensure(); err != nil {
			mw.logger.Error().Err(err).Msg("Invalid request signature")
			http.Error(w, "Invalid request signature", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
