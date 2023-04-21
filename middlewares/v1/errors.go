// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package middlewares

import (
	"context"
	"net/http"

	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/validate"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	v1 "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web/v1"
	"go.uber.org/zap"
)

// errorResponse checks the error and returns the
// appropriate error response and status.
func errorResponse(err error) (v1.ErrorResponse, int) {
	var er v1.ErrorResponse
	var status int
	switch {
	case validate.IsFieldErrors(err):
		fieldErrors := validate.GetFieldErrors(err)
		er = v1.ErrorResponse{
			Error:  "data validation error",
			Fields: fieldErrors.Fields(),
		}
		status = http.StatusBadRequest

	case v1.IsRequestError(err):
		reqErr := v1.GetRequestError(err)
		er = v1.ErrorResponse{
			Error: reqErr.Error(),
		}
		status = reqErr.Status

	default:
		er = v1.ErrorResponse{
			Error: http.StatusText(http.StatusInternalServerError),
		}
		status = http.StatusInternalServerError
	}
	return er, status
}

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			_, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated error.
			if err := handler(ctx, w, r); err != nil {

				// Log the error.
				log.Errorw("ERROR", "message", err)

				// Build out the error response.
				er, status := errorResponse(err)

				// Respond with the error back to the client.
				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}

			// The error has been handled so we can stop propagating it.
			return nil
		}

		return h
	}

	return m
}
