// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"strings"
	"time"
)

type EntryQueryBuilder struct {
	store      *Storage
	feedID     int64
	userID     int64
	timezone   string
	categoryID int64
	status     string
	order      string
	direction  string
	limit      int
	offset     int
	entryID    int64
	gtEntryID  int64
	ltEntryID  int64
	conditions []string
	args       []interface{}
}

func (e *EntryQueryBuilder) WithCondition(column, operator string, value interface{}) *EntryQueryBuilder {
	e.args = append(e.args, value)
	e.conditions = append(e.conditions, fmt.Sprintf("%s %s $%d", column, operator, len(e.args)+1))
	return e
}

func (e *EntryQueryBuilder) WithEntryID(entryID int64) *EntryQueryBuilder {
	e.entryID = entryID
	return e
}

func (e *EntryQueryBuilder) WithEntryIDGreaterThan(entryID int64) *EntryQueryBuilder {
	e.gtEntryID = entryID
	return e
}

func (e *EntryQueryBuilder) WithEntryIDLowerThan(entryID int64) *EntryQueryBuilder {
	e.ltEntryID = entryID
	return e
}

func (e *EntryQueryBuilder) WithFeedID(feedID int64) *EntryQueryBuilder {
	e.feedID = feedID
	return e
}

func (e *EntryQueryBuilder) WithCategoryID(categoryID int64) *EntryQueryBuilder {
	e.categoryID = categoryID
	return e
}

func (e *EntryQueryBuilder) WithStatus(status string) *EntryQueryBuilder {
	e.status = status
	return e
}

func (e *EntryQueryBuilder) WithOrder(order string) *EntryQueryBuilder {
	e.order = order
	return e
}

func (e *EntryQueryBuilder) WithDirection(direction string) *EntryQueryBuilder {
	e.direction = direction
	return e
}

func (e *EntryQueryBuilder) WithLimit(limit int) *EntryQueryBuilder {
	e.limit = limit
	return e
}

func (e *EntryQueryBuilder) WithOffset(offset int) *EntryQueryBuilder {
	e.offset = offset
	return e
}

func (e *EntryQueryBuilder) CountEntries() (count int, err error) {
	defer helper.ExecutionTime(
		time.Now(),
		fmt.Sprintf("[EntryQueryBuilder:CountEntries] userID=%d, feedID=%d, status=%s", e.userID, e.feedID, e.status),
	)

	query := `SELECT count(*) FROM entries e LEFT JOIN feeds f ON f.id=e.feed_id WHERE %s`
	args, condition := e.buildCondition()
	err = e.store.db.QueryRow(fmt.Sprintf(query, condition), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to count entries: %v", err)
	}

	return count, nil
}

func (e *EntryQueryBuilder) GetEntry() (*model.Entry, error) {
	e.limit = 1
	entries, err := e.GetEntries()
	if err != nil {
		return nil, err
	}

	if len(entries) != 1 {
		return nil, nil
	}

	entries[0].Enclosures, err = e.store.GetEnclosures(entries[0].ID)
	if err != nil {
		return nil, err
	}

	return entries[0], nil
}

func (e *EntryQueryBuilder) GetEntries() (model.Entries, error) {
	debugStr := "[EntryQueryBuilder:GetEntries] userID=%d, feedID=%d, categoryID=%d, status=%s, order=%s, direction=%s, offset=%d, limit=%d"
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf(debugStr, e.userID, e.feedID, e.categoryID, e.status, e.order, e.direction, e.offset, e.limit))

	query := `
		SELECT
		e.id, e.user_id, e.feed_id, e.hash, e.published_at at time zone '%s', e.title, e.url, e.author, e.content, e.status,
		f.title as feed_title, f.feed_url, f.site_url, f.checked_at,
		f.category_id, c.title as category_title,
		fi.icon_id
		FROM entries e
		LEFT JOIN feeds f ON f.id=e.feed_id
		LEFT JOIN categories c ON c.id=f.category_id
		LEFT JOIN feed_icons fi ON fi.feed_id=f.id
		WHERE %s %s
	`

	args, conditions := e.buildCondition()
	query = fmt.Sprintf(query, e.timezone, conditions, e.buildSorting())
	// log.Println(query)

	rows, err := e.store.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get entries: %v", err)
	}
	defer rows.Close()

	entries := make(model.Entries, 0)
	for rows.Next() {
		var entry model.Entry
		var iconID interface{}

		entry.Feed = &model.Feed{UserID: e.userID}
		entry.Feed.Category = &model.Category{UserID: e.userID}
		entry.Feed.Icon = &model.FeedIcon{}

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.FeedID,
			&entry.Hash,
			&entry.Date,
			&entry.Title,
			&entry.URL,
			&entry.Author,
			&entry.Content,
			&entry.Status,
			&entry.Feed.Title,
			&entry.Feed.FeedURL,
			&entry.Feed.SiteURL,
			&entry.Feed.CheckedAt,
			&entry.Feed.Category.ID,
			&entry.Feed.Category.Title,
			&iconID,
		)

		if err != nil {
			return nil, fmt.Errorf("Unable to fetch entry row: %v", err)
		}

		if iconID == nil {
			entry.Feed.Icon.IconID = 0
		} else {
			entry.Feed.Icon.IconID = iconID.(int64)
		}

		entry.Feed.ID = entry.FeedID
		entry.Feed.Icon.FeedID = entry.FeedID
		entries = append(entries, &entry)
	}

	return entries, nil
}

func (e *EntryQueryBuilder) buildCondition() ([]interface{}, string) {
	args := []interface{}{e.userID}
	conditions := []string{"e.user_id = $1"}

	if len(e.conditions) > 0 {
		conditions = append(conditions, e.conditions...)
		args = append(args, e.args...)
	}

	if e.categoryID != 0 {
		conditions = append(conditions, fmt.Sprintf("f.category_id=$%d", len(args)+1))
		args = append(args, e.categoryID)
	}

	if e.feedID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.feed_id=$%d", len(args)+1))
		args = append(args, e.feedID)
	}

	if e.entryID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.id=$%d", len(args)+1))
		args = append(args, e.entryID)
	}

	if e.gtEntryID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.id > $%d", len(args)+1))
		args = append(args, e.gtEntryID)
	}

	if e.ltEntryID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.id < $%d", len(args)+1))
		args = append(args, e.ltEntryID)
	}

	if e.status != "" {
		conditions = append(conditions, fmt.Sprintf("e.status=$%d", len(args)+1))
		args = append(args, e.status)
	}

	return args, strings.Join(conditions, " AND ")
}

func (e *EntryQueryBuilder) buildSorting() string {
	var queries []string

	if e.order != "" {
		queries = append(queries, fmt.Sprintf(`ORDER BY "%s"`, e.order))
	}

	if e.direction != "" {
		queries = append(queries, fmt.Sprintf(`%s`, e.direction))
	}

	if e.limit != 0 {
		queries = append(queries, fmt.Sprintf(`LIMIT %d`, e.limit))
	}

	if e.offset != 0 {
		queries = append(queries, fmt.Sprintf(`OFFSET %d`, e.offset))
	}

	return strings.Join(queries, " ")
}

func NewEntryQueryBuilder(store *Storage, userID int64, timezone string) *EntryQueryBuilder {
	return &EntryQueryBuilder{
		store:    store,
		userID:   userID,
		timezone: timezone,
	}
}
