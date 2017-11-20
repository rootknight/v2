// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
)

func (c *Controller) ShowUnreadPage(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	offset := request.GetQueryIntegerParam("offset", 0)

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.DefaultSortingDirection)
	builder.WithOffset(offset)
	builder.WithLimit(NbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	countUnread, err := builder.CountEntries()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("unread", tplParams{
		"user":        user,
		"countUnread": countUnread,
		"entries":     entries,
		"pagination":  c.getPagination(ctx.GetRoute("unread"), countUnread, offset),
		"menu":        "unread",
		"csrf":        ctx.GetCsrfToken(),
	})
}
