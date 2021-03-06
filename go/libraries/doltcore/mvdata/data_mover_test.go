// Copyright 2019 Liquidata, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mvdata

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema/encoding"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table"
)

const (
	schemaFile  = "schema.json"
	mappingFile = "mapping.json"
)

func TestDataMover(t *testing.T) {
	// todo: add expected schema
	tests := []struct {
		sqlSchema   string
		mappingJSON string
		mvOpts      *MoveOptions
	}{
		{
			"",
			"",
			&MoveOptions{
				Operation:   OverwriteOp,
				TableName:   "testable",
				ContOnErr:   false,
				SchFile:     "",
				MappingFile: "",
				PrimaryKey:  "",
				Src:         NewDataLocation("data.csv", ""),
				Dest:        NewDataLocation("data.psv", "psv")},
		},
		/*{
			"",
			"",
			&MoveOptions{
				Operation:   OverwriteOp,
				ContOnErr:   false,
				SchFile:     "",
				MappingFile: "",
				PrimaryKey:  "a",
				Src:         NewDataLocation("data.csv", ""),
				Dest:        NewDataLocation("data.nbf", "")},
		},
		{
			"",
			"",
			&MoveOptions{
				Operation:   OverwriteOp,
				ContOnErr:   false,
				SchFile:     "",
				MappingFile: "",
				PrimaryKey:  "",
				Src:         NewDataLocation("data.nbf", "nbf"),
				Dest:        NewDataLocation("table-name", "")},
		},*/
		{
			"",
			"",
			&MoveOptions{
				Operation:   OverwriteOp,
				TableName:   "table-name",
				ContOnErr:   false,
				SchFile:     "",
				MappingFile: "",
				PrimaryKey:  "a",
				Src:         NewDataLocation("data.csv", ""),
				Dest:        NewDataLocation("table-name", "")},
		},
		{
			`CREATE TABLE table_name (
pk   VARCHAR(120) COMMENT 'tag:0',
value INT          COMMENT 'tag:1',
PRIMARY KEY (pk)
);`,
			`{"a":"pk","b":"value"}`,
			&MoveOptions{
				Operation:   OverwriteOp,
				TableName:   "table_name",
				ContOnErr:   false,
				SchFile:     "",
				MappingFile: "",
				PrimaryKey:  "",
				Src:         NewDataLocation("data.csv", ""),
				Dest:        NewDataLocation("table_name", "")},
		},
	}

	for idx, test := range tests {
		fmt.Println(idx)

		var err error
		_, root, fs := createRootAndFS()

		if test.sqlSchema != "" {
			test.mvOpts.SchFile = schemaFile
			err = fs.WriteFile(schemaFile, []byte(test.sqlSchema))
		}

		if test.mappingJSON != "" {
			test.mvOpts.MappingFile = mappingFile
			err = fs.WriteFile(mappingFile, []byte(test.mappingJSON))
		}

		src := test.mvOpts.Src

		seedWr, err := src.NewCreatingWriter(context.Background(), test.mvOpts, root, fs, true, fakeSchema, nil)

		if err != nil {
			t.Fatal(err.Error())
		}

		imtRd := table.NewInMemTableReader(imt)

		_, _, err = table.PipeRows(context.Background(), imtRd, seedWr, false)
		seedWr.Close(context.Background())
		imtRd.Close(context.Background())

		if err != nil {
			t.Fatal(err)
		}

		encoding.UnmarshalJson(test.sqlSchema)

		dm, crDMErr := NewDataMover(context.Background(), root, fs, test.mvOpts, nil)

		if crDMErr != nil {
			t.Fatal(crDMErr.String())
		}

		var badCount int64
		badCount, err = dm.Move(context.Background())
		assert.Equal(t, int64(0), badCount)

		if err != nil {
			t.Fatal(err)
		}
	}
}
