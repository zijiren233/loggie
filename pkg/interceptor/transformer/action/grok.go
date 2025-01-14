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

package action

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/loggie-io/loggie/pkg/core/api"
	"github.com/loggie-io/loggie/pkg/core/cfg"
	"github.com/loggie-io/loggie/pkg/core/log"
	"github.com/loggie-io/loggie/pkg/util/eventops"
	"github.com/pkg/errors"
)

const (
	grokName     = "grok"
	grokUsageMsg = "usage: grok(key) or grok(key, to)"
)

var DefaultgrokPattern = map[string]string{
	"USERNAME":  "[a-zA-Z0-9._-]+",
	"USER":      "%{USERNAME}",
	"INT":       "(?:[+-]?(?:[0-9]+))",
	"WORD":      "\\b\\w+\\b",
	"UUID":      "[A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12}",
	"IPV4":      "(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)",
	"PATH":      "(?:%{UNIXPATH}|%{WINPATH})",
	"UNIXPATH":  "(/[\\w_%!$@:.,-]?/?)(\\S+)?",
	"WINPATH":   "([A-Za-z]:|\\\\)(?:\\\\[^\\\\?*]*)+",
	"MONTHNUM":  "(?:0?[1-9]|1[0-2])",
	"MONTHDAY":  "(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])",
	"YEAR":      "(\\d\\d){1,2}",
	"DATE_US":   "%{MONTHNUM}[/-]%{MONTHDAY}[/-]%{YEAR}",
	"DATE_EU":   "%{MONTHDAY}[./-]%{MONTHNUM}[./-]%{YEAR}",
	"DATE_CN":   "%{YEAR}[./-]%{MONTHNUM}[./-]%{MONTHDAY}",
	"DATE":      "%{DATE_US}|%{DATE_EU}|%{DATE_CN}",
	"HOUR":      "(?:2[0123]|[01]?[0-9])",
	"MINUTE":    "(?:[0-5][0-9])",
	"SECOND":    "(?:(?:[0-5][0-9]|60)(?:[:.,][0-9]+)?)",
	"TIME":      "([^0-9]?)%{HOUR}:%{MINUTE}(?::%{SECOND})([^0-9]?)",
	"DATESTAMP": "%{DATE}[- ]%{TIME}",
}

func init() {
	RegisterAction(grokName, func(args []string, extra cfg.CommonCfg) (Action, error) {
		return NewGrokAction(args, extra)
	})
}

type grokAction struct {
	key    string
	to     string
	config *grokConfig
	grok   *grok
}

type grokConfig struct {
	Match        string            `yaml:"match,omitempty" validate:"required"`
	IgnoreBlank  *bool             `yaml:"ignoreBlank"`
	PatternPaths []string          `yaml:"patternPaths,omitempty"`
	Pattern      map[string]string `yaml:"pattern,omitempty"`
}

type grok struct {
	p           *regexp.Regexp
	subexpNames []string
	ignoreBlank bool

	patterns     map[string]string
	patternPaths []string
}

func NewGrokAction(args []string, extra cfg.CommonCfg) (*grokAction, error) {
	aCount := len(args)
	if aCount != 1 && aCount != 2 {
		return nil, errors.Errorf("invalid args, %s", grokUsageMsg)
	}
	if aCount == 1 {
		args = append(args, HeaderRoot)
	}

	extraCfg := &grokConfig{}
	if err := cfg.UnpackFromCommonCfg(extra, extraCfg).Do(); err != nil {
		return nil, err
	}

	g := &grokAction{
		key:    args[0],
		to:     args[1],
		config: extraCfg,
	}
	g.initGrok()
	return g, nil
}

func (r *grokAction) initGrok() {
	ignoreBlank := true
	if r.config.IgnoreBlank != nil {
		ignoreBlank = *r.config.IgnoreBlank
	}
	g := NewGrok(r.config.Match, r.config.PatternPaths, ignoreBlank, r.config.Pattern)
	r.grok = g
}

func (r *grokAction) act(e api.Event) error {
	// get field value
	val := eventops.GetString(e, r.key)

	rst := r.grok.grok(val)
	if len(rst) == 0 {
		return errors.Errorf("match group with grok %s is empty", r.config.Match)
	}

	if r.to == HeaderRoot {
		for field, value := range rst {
			eventops.Set(e, field, value)
		}
	} else {
		eventops.Set(e, r.to, rst)
	}

	if r.key != r.to {
		eventops.Del(e, r.key)
	}

	return nil
}

func NewGrok(match string, patternPaths []string, ignoreBlank bool, pattern map[string]string) *grok {
	grok := &grok{
		patternPaths: patternPaths,
		patterns:     DefaultgrokPattern,
		ignoreBlank:  ignoreBlank,
	}
	if len(patternPaths) != 0 {
		grok.loadPatterns()
	}
	if pattern != nil {
		for k, v := range pattern {
			grok.patterns[k] = v
		}
	}
	finalPattern := grok.translateMatchPattern(match)
	p, err := regexp.Compile(finalPattern)
	if err != nil {
		log.Error("could not build grok:%s", err)
	}
	grok.p = p
	grok.subexpNames = p.SubexpNames()

	return grok
}

func (grok *grok) grok(input string) map[string]interface{} {
	rst := make(map[string]interface{})
	for i, substring := range grok.p.FindStringSubmatch(input) {
		if grok.subexpNames[i] == "" {
			continue
		}
		if grok.ignoreBlank && substring == "" {
			continue
		}
		rst[grok.subexpNames[i]] = substring
	}
	return rst
}

func (grok *grok) loadPatterns() {
	for _, patternPath := range grok.patternPaths {
		if strings.HasPrefix(patternPath, "http://") || strings.HasPrefix(patternPath, "https://") {
			resp, err := http.Get(patternPath)
			if err != nil {
				log.Error("load pattern error:%s", err)
				continue
			}
			resp.Body.Close()
			r := bufio.NewReader(resp.Body)
			grok.parseLine(r)
		} else {
			err := grok.parseFiles(patternPath)
			if err != nil {
				log.Error("get files error %v", err)
			}
		}
	}
}

func (grok *grok) parseFiles(filepath string) error {
	fi, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	list, err := f.Readdir(-1)
	defer f.Close()

	if err != nil {
		return err
	}
	for _, l := range list {
		if l.Mode().IsRegular() {
			f, err := os.Open(path.Join(filepath, l.Name()))
			if err != nil {
				log.Error("load pattern error:%s", err)
			}
			r := bufio.NewReader(f)
			grok.parseLine(r)
		}
	}
	return nil
}

func (grok *grok) parseLine(r *bufio.Reader) {
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("read patterns error:%s", err)
			continue
		}
		if isPrefix == true {
			log.Error("readline prefix")
			continue
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		kv := strings.SplitN(string(line), " ", 2)
		grok.patterns[kv[0]] = strings.TrimSpace(kv[1])
	}
}

func (grok *grok) translateMatchPattern(s string) string {
	p, err := regexp.Compile(`%{\w+?(:\w+?)?}`)
	if err != nil {
		log.Error("could not translate match pattern: %v", err.Error())
		return ""
	}
	var r string
	for {
		r = p.ReplaceAllStringFunc(s, grok.replaceFunc)
		if r == s {
			return r
		}
		s = r
	}
}

func (grok *grok) replaceFunc(s string) string {
	p, err := regexp.Compile(`%{(\w+?)(?::(\w+?))?}`)
	if err != nil {
		log.Error("could not replace match func: %v", err.Error())
		return ""
	}
	rst := p.FindAllStringSubmatch(s, -1)
	if len(rst) != 1 {
		log.Error("sub match in `%s` != 1", s)
		return ""
	}
	if pattern, ok := grok.patterns[rst[0][1]]; ok {
		if rst[0][2] == "" {
			return fmt.Sprintf("(%s)", pattern)
		} else {
			return fmt.Sprintf("(?P<%s>%s)", rst[0][2], pattern)
		}
	} else {
		log.Error("`%s` could not be found", rst[0][1])
		return ""
	}
}
