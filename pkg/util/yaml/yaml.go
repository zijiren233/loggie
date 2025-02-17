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

package yaml

import (
	"fmt"

	goccyyaml "github.com/goccy/go-yaml"
	"gopkg.in/yaml.v2"
)

func Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

func UnmarshalWithPrettyError(in []byte, out interface{}) error {
	err := Unmarshal(in, out)
	if err != nil {
		prettyErr := goccyyaml.Unmarshal(in, out)
		if prettyErr != nil {
			err = fmt.Errorf("%w\n; %s", prettyErr, err)
		}
	}
	return err
}

func Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}
