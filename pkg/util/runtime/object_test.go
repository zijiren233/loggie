/*
Copyright 2021 Loggie Authors

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

package runtime

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var data = map[string]interface{}{
	"a": "b",
	"c": 1,
	"d": map[string]interface{}{
		"e": "f",
		"g": 2,
	},
}

func TestObject_Get(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "ok-string",
			fields: fields{
				data: data,
			},
			args: args{
				key: "a",
			},
			want: &Object{
				data: "b",
			},
		},
		{
			name: "ok-map",
			fields: fields{
				data: data,
			},
			args: args{
				key: "d",
			},
			want: &Object{
				data: map[string]interface{}{
					"e": "f",
					"g": 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			got := obj.Get(tt.args.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestObject_GetPaths(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		paths []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "ok",
			fields: fields{
				data: data,
			},
			args: args{
				paths: []string{"d", "e"},
			},
			want: &Object{
				data: "f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			if got := obj.GetPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_DelPaths(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		paths []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "ok",
			fields: fields{
				data: data,
			},
			args: args{
				paths: []string{"d", "e"},
			},
			want: &Object{
				data: map[string]interface{}{
					"a": "b",
					"c": 1,
					"d": map[string]interface{}{
						"g": 2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			obj.DelPaths(tt.args.paths)

			if !reflect.DeepEqual(obj, tt.want) {
				t.Errorf("DelPaths() = %v, want %v", obj, tt.want)
			}
		})
	}
}

func TestObject_SetPaths(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		paths []string
		val   interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "ok",
			fields: fields{
				data: map[string]interface{}{
					"a": "b",
					"d": map[string]interface{}{
						"e": "f",
						"g": 2,
					},
				},
			},
			args: args{
				paths: []string{"d", "h"},
				val:   "k",
			},
			want: &Object{
				data: map[string]interface{}{
					"a": "b",
					"d": map[string]interface{}{
						"e": "f",
						"g": 2,
						"h": "k",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			obj.SetPaths(tt.args.paths, tt.args.val)
			if !reflect.DeepEqual(obj, tt.want) {
				t.Errorf("SetPaths() = %v, want %v", obj, tt.want)
			}
		})
	}
}

func TestObject_GetPath(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		query string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "ok-sin",
			fields: fields{
				data: map[string]interface{}{
					"a": "b",
					"c": 1,
					"d": map[string]interface{}{
						"e": "f",
						"g": 2,
					},
				},
			},
			args: args{
				query: "c",
			},
			want: &Object{
				data: 1,
			},
		},
		{
			name: "ok-paths",
			fields: fields{
				data: map[string]interface{}{
					"a": "b",
					"c": 1,
					"d": map[string]interface{}{
						"e": "f",
						"g": 2,
					},
				},
			},
			args: args{
				query: "d.e",
			},
			want: &Object{
				data: "f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			if got := obj.GetPath(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_FlatKeyValue(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "ok",
			fields: fields{
				data: map[string]interface{}{
					"a": "f",
					"b": map[string]interface{}{
						"c": "g",
					},
				},
			},
			want: map[string]interface{}{
				"a":   "f",
				"b_c": "g",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			got, err := obj.FlatKeyValue("_")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestObject_ConvertKeys(t *testing.T) {
	type fields struct {
		data interface{}
	}
	type args struct {
		keyFunc convertKeyFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Object
	}{
		{
			name: "update keys ok",
			fields: fields{
				data: map[string]interface{}{
					"a": "b",
					"c": "d",
					"e": map[string]interface{}{
						"f": "g",
					},
				},
			},
			args: args{
				keyFunc: func(key string) string {
					if key == "a" {
						return "aa"
					}
					if key == "f" {
						return "ff"
					}
					return ""
				},
			},
			want: &Object{
				data: map[string]interface{}{
					"aa": "b",
					"c":  "d",
					"e": map[string]interface{}{
						"ff": "g",
					},
				},
			},
		},
		{
			name: "regex keys ok",
			fields: fields{
				data: map[string]interface{}{
					"a": "b",
					"e": map[string]interface{}{
						"foo.bar/test": "g",
					},
				},
			},
			args: args{
				keyFunc: func(key string) string {
					reg := regexp.MustCompile("foo.bar/(.*)")
					matched := reg.FindStringSubmatch(key)
					if len(matched) > 0 {
						return reg.ReplaceAllString(key, "pre-${1}")
					}
					return ""
				},
			},
			want: &Object{
				data: map[string]interface{}{
					"a": "b",
					"e": map[string]interface{}{
						"pre-test": "g",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{
				data: tt.fields.data,
			}
			err := obj.ConvertKeys(tt.args.keyFunc)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, obj)
		})
	}
}
