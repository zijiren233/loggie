/*
Copyright 2022 Loggie Authors

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

package condition

import (
	"testing"

	"github.com/loggie-io/loggie/pkg/core/api"
	"github.com/loggie-io/loggie/pkg/core/event"
	"github.com/stretchr/testify/assert"
)

func TestIsNull_Check(t *testing.T) {
	assertions := assert.New(t)

	type fields struct {
		field string
	}
	type args struct {
		e api.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "fields is null",
			fields: fields{
				field: "a.c",
			},
			args: args{
				e: event.NewEvent(map[string]interface{}{
					"a": map[string]interface{}{
						"b": "xxx",
						"c": nil,
					},
				}, []byte("this is body")),
			},
			want: false,
		},
		{
			name: "fields is not exist",
			fields: fields{
				field: "a.c",
			},
			args: args{
				e: event.NewEvent(map[string]interface{}{
					"a": map[string]interface{}{
						"b": "xxx",
					},
				}, []byte("this is body")),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &Exist{
				field: tt.fields.field,
			}

			got := et.Check(tt.args.e)
			assertions.Equal(tt.want, got, "check failed")
		})
	}
}
