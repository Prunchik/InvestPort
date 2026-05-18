package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePaginationQuery_Valid(t *testing.T) {
	tests := []struct {
		name     string
		query    url.Values
		expected PaginationQuery
	}{
		{
			name:     "no params - defaults",
			query:    url.Values{},
			expected: PaginationQuery{Offset: 0, Limit: 20},
		},
		{
			name:     "only offset",
			query:    url.Values{"offset": []string{"10"}},
			expected: PaginationQuery{Offset: 10, Limit: 20},
		},
		{
			name:     "only limit",
			query:    url.Values{"limit": []string{"50"}},
			expected: PaginationQuery{Offset: 0, Limit: 50},
		},
		{
			name:     "both valid",
			query:    url.Values{"offset": []string{"5"}, "limit": []string{"30"}},
			expected: PaginationQuery{Offset: 5, Limit: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tt.query.Encode()}}
			result, err := parsePaginationQuery(r)
			assert.NoError(t, err)
			assert.Equal(t, &tt.expected, result)
		})
	}
}

func TestParsePaginationQuery_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
	}{
		{
			name:  "invalid offset",
			query: url.Values{"offset": []string{"abc"}},
		},
		{
			name:  "negative offset",
			query: url.Values{"offset": []string{"-5"}},
		},
		{
			name:  "invalid limit",
			query: url.Values{"limit": []string{"xyz"}},
		},
		{
			name:  "zero limit",
			query: url.Values{"limit": []string{"0"}},
		},
		{
			name:  "limit too big",
			query: url.Values{"limit": []string{"150"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tt.query.Encode()}}
			_, err := parsePaginationQuery(r)
			assert.Error(t, err)
		})
	}
}

func TestParseHistoryQuery_Valid(t *testing.T) {
	tests := []struct {
		name     string
		query    url.Values
		expected HistoryQuery
	}{
		{
			name: "defaults only",
			query: url.Values{
				"offset": []string{"0"},
				"limit":  []string{"20"},
			},
			expected: HistoryQuery{
				PaginationQuery: PaginationQuery{Offset: 0, Limit: 20},
				Interval:        "hour",
				Mode:            "last",
			},
		},
		{
			name: "valid interval and mode",
			query: url.Values{
				"interval": []string{"day"},
				"mode":     []string{"avg"},
				"offset":   []string{"10"},
				"limit":    []string{"50"},
			},
			expected: HistoryQuery{
				PaginationQuery: PaginationQuery{Offset: 10, Limit: 50},
				Interval:        "day",
				Mode:            "avg",
			},
		},
		{
			name: "unsupported interval falls back to hour",
			query: url.Values{
				"interval": []string{"month"},
				"mode":     []string{"last"},
			},
			expected: HistoryQuery{
				PaginationQuery: PaginationQuery{Offset: 0, Limit: 20},
				Interval:        "hour",
				Mode:            "last",
			},
		},
		{
			name: "unsupported mode falls back to last",
			query: url.Values{
				"mode": []string{"min"},
			},
			expected: HistoryQuery{
				PaginationQuery: PaginationQuery{Offset: 0, Limit: 20},
				Interval:        "hour",
				Mode:            "last",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tt.query.Encode()}}
			result, err := parseHistoryQuery(r)
			assert.NoError(t, err)
			assert.Equal(t, &tt.expected, result)
		})
	}
}

func TestParseHistoryQuery_InvalidPagination(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
	}{
		{
			name:  "invalid offset",
			query: url.Values{"offset": []string{"invalid"}},
		},
		{
			name:  "invalid limit",
			query: url.Values{"limit": []string{"invalid"}},
		},
		{
			name:  "negative offset",
			query: url.Values{"offset": []string{"-5"}},
		},
		{
			name:  "zero limit",
			query: url.Values{"limit": []string{"0"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tt.query.Encode()}}
			_, err := parseHistoryQuery(r)
			assert.Error(t, err)
		})
	}
}
