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
		// 1. Read the entire body once
		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			mw.logger.Error().Err(err).Msg("Failed to read request body")
			http.Error(w, "Failed to read request", http.StatusInternalServerError)

			return
		}

		// Always close the original body when done
		r.Body.Close()

		// 2. Log body content if needed
		mw.logger.Info().Str("body", string(bodyBytes)).Msg("Request body")

		// 3. Perform Slack signature verification
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

		// 4. Set the body back for form processing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 5. Parse the form to make form values available
		if err := r.ParseForm(); err != nil {
			mw.logger.Error().Err(err).Msg("Failed to parse form")
			http.Error(w, "Failed to parse form", http.StatusBadRequest)

			return
		}

		// 6. Set the body back again for possible binding in Echo
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}
