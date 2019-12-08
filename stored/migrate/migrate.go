/*
Copyright 2019 The Predictive Horizontal Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package migrate handles applying migrations to the sqlite3 db for storing evaluations
package migrate

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Driver for loading evaluations from file system
	_ "github.com/mattn/go-sqlite3"                      // Driver for sqlite3 database
)

// Migrate applies the database migrations to the sqlite3 db
func Migrate(migrationSource string, db *sql.DB) error {
	// Get DB driver
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	// Load migrations
	m, err := migrate.NewWithDatabaseInstance(migrationSource, "evaluations", driver)
	if err != nil {
		return err
	}

	// Apply migrations
	err = m.Up()
	if err != nil {
		return err
	}

	return nil
}
