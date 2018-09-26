package auth

import (
	"regexp"
	"strings"

	"github.com/danesparza/iamserver/compiler"
	"github.com/danesparza/iamserver/data"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

var (
	regexpStartDelimeter byte
	regexpEndDelimeter   byte
)

// NewRegexpMatcher creates and returns a new RegexpMatcher
func NewRegexpMatcher(size int) *RegexpMatcher {
	if size <= 0 {
		size = 512
	}

	// golang-lru only returns an error if the cache's size is 0. This, we can safely ignore this error.
	cache, _ := lru.New(size)
	return &RegexpMatcher{
		Cache: cache,
	}
}

// RegexpMatcher represents a regular expression matcher
type RegexpMatcher struct {
	*lru.Cache

	C map[string]*regexp.Regexp
}

func (m *RegexpMatcher) get(pattern string) *regexp.Regexp {
	if val, ok := m.Cache.Get(pattern); !ok {
		return nil
	} else if reg, ok := val.(*regexp.Regexp); !ok {
		return nil
	} else {
		return reg
	}
}

func (m *RegexpMatcher) set(pattern string, reg *regexp.Regexp) {
	m.Cache.Add(pattern, reg)
}

// Matches a needle with an array of regular expressions and returns true if a match was found.
func (m *RegexpMatcher) Matches(p data.Policy, haystack []string, needle string) (bool, error) {
	var reg *regexp.Regexp
	var err error
	for _, h := range haystack {

		// This means that the current haystack item does not contain a regular expression
		if strings.Count(h, string(regexpStartDelimeter)) == 0 {
			// If we have a simple string match, we've got a match!
			if h == needle {
				return true, nil
			}

			// Not string match, but also no regexp, continue with next haystack item
			continue
		}

		if reg = m.get(h); reg != nil {
			if reg.MatchString(needle) {
				return true, nil
			}
			continue
		}

		reg, err = compiler.CompileRegex(h, regexpStartDelimeter, regexpEndDelimeter)
		if err != nil {
			return false, errors.WithStack(err)
		}

		m.set(h, reg)
		if reg.MatchString(needle) {
			return true, nil
		}
	}
	return false, nil
}

func init() {
	regexpStartDelimeter = '<'
	regexpEndDelimeter = '>'
}
