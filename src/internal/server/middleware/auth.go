package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"go.pkg.littleman.co/library/internal/problems"
	"golang.org/x/oauth2"
)

// CookieNameReturnURL the URL that users will be redirected to once authentication is complete
const CookieNameReturnURL = "return-url"

// CookieAuthentication the authentication token that users will be verified against
const CookieAuthentication = "authentication"

// OIDCClaimSet is a set of claims that must match collectively for the autentication to continue
type OIDCClaimSet map[string]string

var problem = &problems.Factory{
	URITemplate: "https://github.com/littlemanco/library/tree/master/docs/errors/__ID__.md",
}

// OidcAuth is an object that creates the OIDC Middleware primitive
type OidcAuth struct {
	// Public
	OIDCProvider *oidc.Provider
	OAuth2       *oauth2.Config
	RedirectURL  *url.URL
	Claims       []OIDCClaimSet

	// State is a random
	// State        string
	// Todo: Inject telemetry
}

// OIDCAuthConfiguration is a function that modifies OIDC Auth behaviour
type OIDCAuthConfiguration func(o *OidcAuth) error

// Write the URL users should return to to a persistent store
func writeToURL(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameReturnURL,
		Value:    r.RequestURI,
		Path:     "/",
		Expires:  time.Now().Add(60 * time.Minute),
		HttpOnly: true,
	})
}

// Read the URL users should return to the persistent store back
func readAndClearToURL(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(CookieNameReturnURL)

	if err != nil {
		return "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameReturnURL,
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
	options ...OIDCAuthConfiguration,
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

	// Validate constructed object
	if len(auth.Claims) == 0 {
		return nil, problem.WithTitle("Missing OIDC Claims")
	}

	return auth, nil
}

// WithClaimSet allows requiring specific characteristics of the OIDC to verify against
func WithClaimSet(set OIDCClaimSet) func(o *OidcAuth) error {
	return func(o *OidcAuth) error {
		// Validate there are actually OIDC claims
		if len(set) == 0 {
			return problem.WithTitle("Empty set of OIDC Claims supplied")
		}

		// Add the required claims for later analysis
		o.Claims = append(o.Claims, set)

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

	// Unwrap list
	for _, s := range o.Claims {
		// Test if any claims are valid
		isValidForClaimSet := true
		for k, v := range s {
			val, ok := claims[k]

			// Bail early if the claim is not there
			if !ok {
				isValidForClaimSet = false
				break
			}

			// Bail if an expected claim does not match
			if val != v {
				isValidForClaimSet = false
				break
			}
		}

		// Only a single match needs to be valid. If it is, exit with success.
		if isValidForClaimSet {
			return nil
		}
	}

	return problem.WithTitleAudience("User Missing Valid Claim Set", []int{problems.AudienceConsumer})
}
