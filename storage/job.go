// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"log"
	"time"
)

const maxParsingError = 3

func (s *Storage) GetJobs(batchSize int) []model.Job {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("storage.GetJobs[%d]", batchSize))

	var jobs []model.Job
	query := `SELECT
		id, user_id
		FROM feeds
		WHERE parsing_error_count < $1
		ORDER BY checked_at ASC LIMIT %d`

	rows, err := s.db.Query(fmt.Sprintf(query, batchSize), maxParsingError)
	if err != nil {
		log.Println("Unable to fetch feed jobs:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID); err != nil {
			log.Println("Unable to fetch feed job:", err)
			break
		}

		jobs = append(jobs, job)
	}

	return jobs
}
