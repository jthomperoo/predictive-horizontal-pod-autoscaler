/*
Copyright 2021 The Predictive Horizontal Pod Autoscaler Authors.

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

// Package stored provides interfacing methods for updating/retrieving data from the local sqlite3 database
package stored

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
)

// DBEvaluation is an evaluation that can be stored in a SQL db, it is exactly the same as the evaluation but just with
// a Scan method implementation for marshalling from JSON
type DBEvaluation evaluate.Evaluation

// Scan is used when reading the DBEvaluation field in from the SQL db, marshalling from a JSON string
func (d *DBEvaluation) Scan(src interface{}) error {
	strValue, ok := src.(string)

	if !ok {
		return fmt.Errorf("evaluation field must be a string, got %T instead", src)
	}

	return json.Unmarshal([]byte(strValue), d)
}

// Evaluation is an evaluation stored in the database, with an ID and a timestamp
type Evaluation struct {
	ID         int          `db:"id" json:"id"`
	Created    time.Time    `db:"created" json:"created"`
	Evaluation DBEvaluation `db:"val" json:"val"`
}

// Model is a model stored in the database, with an ID and the number of intervals since it was last used
type Model struct {
	ID              int    `db:"id" json:"id"`
	Name            string `db:"model_name" json:"model_name"`
	IntervalsPassed int    `db:"intervals_passed" json:"intervals_passed"`
}

// Storer allows updating/retrieving evaluations from a data source
type Storer interface {
	GetEvaluation(model string) ([]*Evaluation, error)
	AddEvaluation(model string, evaluation *evaluate.Evaluation) error
	RemoveEvaluation(id int) error
	GetModel(model string) (*Model, error)
	UpdateModel(model string, intervalsPassed int) error
}

// LocalStore is the implementation of a Store for updating/retrieving evaluations from a SQL db
type LocalStore struct {
	DB *sql.DB
}

// GetEvaluation returns all evaluations associated with a model
func (s *LocalStore) GetEvaluation(model string) ([]*Evaluation, error) {
	rows, err := s.DB.Query("SELECT evaluation.id, evaluation.created, evaluation.val FROM evaluation, model WHERE evaluation.model_id = model.id AND model.model_name = ?;", model)
	if err != nil {
		return nil, err
	}

	var saved []*Evaluation
	for rows.Next() {
		evaluation := Evaluation{}
		err = rows.Scan(&evaluation.ID, &evaluation.Created, &evaluation.Evaluation)
		if err != nil {
			return nil, err
		}
		saved = append(saved, &evaluation)
	}

	return saved, nil
}

// AddEvaluation inserts a new evaluation associated with a model
func (s *LocalStore) AddEvaluation(model string, evaluation *evaluate.Evaluation) error {
	// Get the model for the evaluation
	modelObj, err := s.GetModel(model)
	if err != nil {
		return err
	}

	// Convert evaluation into JSON
	evaluationJSON, err := json.Marshal(evaluation)
	if err != nil {
		// Should not occur, panic
		log.Panic(err)
	}

	// Start db transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	// Prepare insert statement
	stmt, err := tx.Prepare("INSERT INTO evaluation(model_id, val, created) VALUES(?, ?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Execute statement
	stmt.Exec(modelObj.ID, string(evaluationJSON), time.Now().UTC().Unix())
	// Execute transaction, return any error
	return tx.Commit()
}

// RemoveEvaluation deletes an evaluation by the ID provided
func (s *LocalStore) RemoveEvaluation(id int) error {
	_, err := s.DB.Exec("DELETE FROM evaluation WHERE id = $1;", id)
	return err
}

// GetModel gets the model with the name provided
func (s *LocalStore) GetModel(model string) (*Model, error) {
	row := s.DB.QueryRow("SELECT id, model_name, intervals_passed FROM model WHERE model_name = ?;", model)

	result := Model{}
	err := row.Scan(&result.ID, &result.Name, &result.IntervalsPassed)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateModel updates the model in the DB with the name provided, setting the intervals_passed value
func (s *LocalStore) UpdateModel(model string, intervalsPassed int) error {
	modelObj, err := s.GetModel(model)
	if err == sql.ErrNoRows {
		// Start db transaction
		tx, err := s.DB.Begin()
		if err != nil {
			return err
		}
		// Prepare insert statement
		stmt, err := tx.Prepare("INSERT INTO model(model_name, intervals_passed) VALUES(?, ?);")
		if err != nil {
			return err
		}
		defer stmt.Close()
		// Execute statement
		stmt.Exec(model, intervalsPassed)
		// Execute transaction, return any error
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	// Start db transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	// Prepare insert statement
	stmt, err := tx.Prepare("UPDATE model SET intervals_passed = ? WHERE model.id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Execute statement
	stmt.Exec(intervalsPassed, modelObj.ID)
	// Execute transaction, return any error
	return tx.Commit()
}
