package dto_test

import (
	"strings"
	"testing"

	"tex/src/internal/dto"

	"github.com/stretchr/testify/assert"
)

func TestSubmitRequest_Validate_Valid(t *testing.T) {
	req := dto.SubmitRequest{
		FullName:   "Somchai Jaidee",
		Email:      "somchai@example.com",
		TopicTitle: "Go Generics in Practice",
	}
	assert.NoError(t, req.Validate())
}

func TestSubmitRequest_Validate_EmptyFullName(t *testing.T) {
	req := dto.SubmitRequest{FullName: "", Email: "a@b.com", TopicTitle: "Topic"}
	assert.EqualError(t, req.Validate(), "full_name is required")
}

func TestSubmitRequest_Validate_WhitespaceFullName(t *testing.T) {
	req := dto.SubmitRequest{FullName: "   ", Email: "a@b.com", TopicTitle: "Topic"}
	assert.EqualError(t, req.Validate(), "full_name is required")
}

func TestSubmitRequest_Validate_FullNameTooLong(t *testing.T) {
	req := dto.SubmitRequest{FullName: strings.Repeat("a", 256), Email: "a@b.com", TopicTitle: "Topic"}
	assert.EqualError(t, req.Validate(), "full_name must not exceed 255 characters")
}

func TestSubmitRequest_Validate_InvalidEmail(t *testing.T) {
	cases := []string{"notanemail", "missing@", "@nodomain", "spaces in@email.com", ""}
	for _, email := range cases {
		req := dto.SubmitRequest{FullName: "Name", Email: email, TopicTitle: "Topic"}
		assert.EqualError(t, req.Validate(), "email is invalid", "expected error for email %q", email)
	}
}

func TestSubmitRequest_Validate_ValidEmailFormats(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"user+tag@example.co.th",
		"user.name@sub.domain.com",
	}
	for _, email := range validEmails {
		req := dto.SubmitRequest{FullName: "Name", Email: email, TopicTitle: "Topic"}
		assert.NoError(t, req.Validate(), "expected no error for email %q", email)
	}
}

func TestSubmitRequest_Validate_EmptyTopicTitle(t *testing.T) {
	req := dto.SubmitRequest{FullName: "Name", Email: "a@b.com", TopicTitle: ""}
	assert.EqualError(t, req.Validate(), "topic_title is required")
}

func TestSubmitRequest_Validate_TopicTitleTooLong(t *testing.T) {
	req := dto.SubmitRequest{FullName: "Name", Email: "a@b.com", TopicTitle: strings.Repeat("x", 256)}
	assert.EqualError(t, req.Validate(), "topic_title must not exceed 255 characters")
}
