// internal/common/response.go
package common

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Total       int64 `json:"total"`
	Limit       int   `json:"limit"`
	Offset      int   `json:"offset"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
}

// PagedResponse extends BaseResponse with pagination information
type PagedResponse[T any] struct {
	IsSuccess  bool           `json:"is_success"`
	Message    string         `json:"message,omitempty"`
	Data       T              `json:"data,omitempty"`
	Error      string         `json:"error,omitempty"`
	Pagination PaginationInfo `json:"pagination,omitempty"`
}

// BaseResponse is a generic wrapper for an API Response.
type BaseResponse[T any] struct {
	IsSuccess bool   `json:"is_success"`
	Message   string `json:"message,omitempty"`
	Data      T      `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}

func NewSuccess[T any](data T, message string) *BaseResponse[T] {
	return &BaseResponse[T]{
		IsSuccess: true,
		Message:   message,
		Data:      data,
	}
}

func NewPagedSuccess[T any](data T, message string, pagination PaginationInfo) *PagedResponse[T] {
	return &PagedResponse[T]{
		IsSuccess:  true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

func NewError(message string, err error) *BaseResponse[string] {
	return &BaseResponse[string]{
		IsSuccess: false,
		Message:   message,
		Error:     err.Error(),
	}
}
