// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/route"
	"github.com/miniflux/miniflux2/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Context contains helper functions related to the current request.
type Context struct {
	writer  http.ResponseWriter
	request *http.Request
	store   *storage.Storage
	router  *mux.Router
	user    *model.User
}

// IsAdminUser checks if the logged user is administrator.
func (c *Context) IsAdminUser() bool {
	if v := c.request.Context().Value("IsAdminUser"); v != nil {
		return v.(bool)
	}
	return false
}

// GetUserTimezone returns the timezone used by the logged user.
func (c *Context) GetUserTimezone() string {
	if v := c.request.Context().Value("UserTimezone"); v != nil {
		return v.(string)
	}
	return "UTC"
}

// IsAuthenticated returns a boolean if the user is authenticated.
func (c *Context) IsAuthenticated() bool {
	if v := c.request.Context().Value("IsAuthenticated"); v != nil {
		return v.(bool)
	}
	return false
}

// GetUserID returns the UserID of the logged user.
func (c *Context) GetUserID() int64 {
	if v := c.request.Context().Value("UserId"); v != nil {
		return v.(int64)
	}
	return 0
}

// GetLoggedUser returns all properties related to the logged user.
func (c *Context) GetLoggedUser() *model.User {
	if c.user == nil {
		var err error
		c.user, err = c.store.GetUserById(c.GetUserID())
		if err != nil {
			log.Fatalln(err)
		}

		if c.user == nil {
			log.Fatalln("Unable to find user from context")
		}
	}

	return c.user
}

// GetUserLanguage get the locale used by the current logged user.
func (c *Context) GetUserLanguage() string {
	user := c.GetLoggedUser()
	return user.Language
}

// GetCsrfToken returns the current CSRF token.
func (c *Context) GetCsrfToken() string {
	if v := c.request.Context().Value("CsrfToken"); v != nil {
		return v.(string)
	}

	log.Println("No CSRF token in context!")
	return ""
}

// GetRoute returns the path for the given arguments.
func (c *Context) GetRoute(name string, args ...interface{}) string {
	return route.GetRoute(c.router, name, args...)
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request, store *storage.Storage, router *mux.Router) *Context {
	return &Context{writer: w, request: r, store: store, router: router}
}
