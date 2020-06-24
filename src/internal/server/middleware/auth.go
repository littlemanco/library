package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// CookieReturnURL the URL that users will be redirected to once authentication is complete
const CookieReturnURL = "return-url"

// CookieAuthentication the authentication token that users will be verified against
const CookieAuthentication = "authentication"

// OidcAuth is an object that creates the OIDC Middleware primitive
type OidcAuth struct {
	OIDCProvider *oidc.Provider
	OAuth2       *oauth2.Config
	RedirectURL  *url.URL
	Claims       map[string]string

	// State is a random
	// State        string
	// Todo: Inject telemetry
}

// Write the URL users should return to to a persistent store
func writeToURL(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieReturnURL,
		Value:    r.RequestURI,
		Path:     "/",
		Expires:  time.Now().Add(60 * time.Minute),
		HttpOnly: true,
	})
}

// Read the URL users should return to the persistent store back
func readAndClearToURL(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(CookieAuthentication)

	if err != nil {
		return "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieReturnURL,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-60 * time.Minute),
		HttpOnly: true,
	})

	return c.Value
}

// NewOidcAuth returns the OIDC Middleware
func NewOidcAuth(
	provider string,
	clientID string,
	clientSecret string,
	redirectURL *url.URL,
	options ...func(o *OidcAuth) error,
) (*OidcAuth, error) {
	p, err := oidc.NewProvider(context.Background(), provider)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to set up oidc middleware")
	}

	auth := &OidcAuth{
		OIDCProvider: p,
		OAuth2: &oauth2.Config{
			Endpoint:     p.Endpoint(),
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL.String(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		RedirectURL: redirectURL,
	}

	for _, o := range options {
		if e := o(auth); e != nil {
			return nil, errors.Wrap(e, "Unable to set up oidc middleware")
		}
	}

	return auth, nil
}

// WithClaims allows requiring specific characteristics of the OIDC to verify against
func WithClaims(claims map[string]string) func(o *OidcAuth) error {
	return func(o *OidcAuth) error {
		// Add the required claims for later analysis
		o.Claims = claims

		for k := range claims {
			o.OAuth2.Scopes = append(o.OAuth2.Scopes, k)
		}

		return nil
	}
}

// Middleware is the actual middleware function to append to routes.
func (o *OidcAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is a callback, route it to the callback handler
		if r.URL.Path == o.RedirectURL.Path {
			o.CallbackHandler(w, r)

			return
		}

		// If the user is not authenticated, redirect them to the place they need to go for auth.
		token, err := r.Cookie(CookieAuthentication)

		// if there is no authentication cookie, Redirect the user to the place to login
		if err != nil {
			// Store the previous URL so users can be sent back once they're authenticated
			writeToURL(w, r)
			http.Redirect(w, r, o.OAuth2.AuthCodeURL("TODO"), http.StatusFound)

			return
		}

		if err := o.verify(token.Value); err != nil {
			http.SetCookie(
				w,
				&http.Cookie{
					Name:    CookieAuthentication,
					Path:    "/",
					Expires: time.Now().Add(-60 * time.Minute),
				},
			)

			http.Error(w, fmt.Sprintf("Unauthorized: %s. Refresh to sign in again.", err.Error()), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CallbackHandler is the handler for redirect requests.
func (o *OidcAuth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Todo: Verify State
	if r.URL.Query().Get("state") != "TODO" {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := o.OAuth2.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "Failed to get token from return: "+err.Error(), http.StatusInternalServerError)
	}

	// Store the token in a cookie
	// Todo: Replace with https://github.com/gorilla/securecookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAuthentication,
		Value:    rawIDToken,
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect the user back to the previously defined URL
	http.Redirect(w, r, readAndClearToURL(w, r), http.StatusFound)
}

func (o *OidcAuth) verify(token string) error {
	claims := map[string]interface{}{}

	verifier := o.OIDCProvider.Verifier(&oidc.Config{ClientID: o.OAuth2.ClientID})

	t, err := verifier.Verify(context.Background(), token)

	if err != nil {
		return errors.Wrap(err, "unable to verify user")
	}

	if err := t.Claims(&claims); err != nil {
		return errors.Wrap(err, "unable to verify user")
	}

	for k, v := range o.Claims {
		val, ok := claims[k]

		// Bail early if the claim is not there
		if !ok {
			return fmt.Errorf("required claim “%s” not present", k)
		}

		if val != v {
			return fmt.Errorf("required claim “%s” does not match the requirement", val)
		}
	}

	return nil
}
