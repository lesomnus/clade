package registry

import (
	"fmt"
)

type ErrorCode struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Detail  *string `json:"detail"`
}

type ErrorCodes struct {
	Errors []ErrorCode `json:"errors"`
}

func NewErrorCodes(ec ...ErrorCode) *ErrorCodes {
	return &ErrorCodes{Errors: ec}
}

func (e *ErrorCodes) Error() string {
	return fmt.Sprintf("%v", e.Errors)
}

// TODO use distribution/.../errcode
var (
	ErrCodeBlobUnknown         = ErrorCode{Code: "BLOB_UNKNOWN", Message: "blob unknown to registry"}
	ErrCodeBlobUploadInvalid   = ErrorCode{Code: "BLOB_UPLOAD_INVALID", Message: "blob upload invalid"}
	ErrCodeBlobUploadUnknown   = ErrorCode{Code: "BLOB_UPLOAD_UNKNOWN", Message: "blob upload unknown to registry"}
	ErrCodeDigestInvalid       = ErrorCode{Code: "DIGEST_INVALID", Message: "provided digest did not match uploaded content"}
	ErrCodeManifestBlobUnknown = ErrorCode{Code: "MANIFEST_BLOB_UNKNOWN", Message: "manifest references a manifest or blob unknown to registry"}
	ErrCodeManifestInvalid     = ErrorCode{Code: "MANIFEST_INVALID", Message: "manifest invalid"}
	ErrCodeManifestUnknown     = ErrorCode{Code: "MANIFEST_UNKNOWN", Message: "manifest unknown to registry"}
	ErrCodeNameInvalid         = ErrorCode{Code: "NAME_INVALID", Message: "invalid repository name"}
	ErrCodeNameUnknown         = ErrorCode{Code: "NAME_UNKNOWN", Message: "repository name not known to registry"}
	ErrCodeSizeInvalid         = ErrorCode{Code: "SIZE_INVALID", Message: "provided length did not match content length"}
	ErrCodeUnauthorized        = ErrorCode{Code: "UNAUTHORIZED", Message: "authentication required"}
	ErrCodeDenied              = ErrorCode{Code: "DENIED", Message: "requested access to the resource is denied"}
	ErrCodeUnsupported         = ErrorCode{Code: "UNSUPPORTED", Message: "the operation is unsupported"}
	ErrCodeTooManyRequests     = ErrorCode{Code: "TOOMANYREQUESTS", Message: "too many requests"}
)
