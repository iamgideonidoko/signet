package validator

import (
"fmt"
"regexp"
"strings"

"github.com/iamgideonidoko/signet/internal/models"
)

var (
hashRegex     = regexp.MustCompile(`^[a-fA-F0-9]{8,128}$`)
timezoneRegex = regexp.MustCompile(`^[A-Za-z]+/[A-Za-z_]+$`)
)

type ValidationError struct {
Field   string
Message string
}

func (e *ValidationError) Error() string {
return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Validator struct {
errors []ValidationError
}

func New() *Validator {
return &Validator{
errors: make([]ValidationError, 0),
}
}

func (v *Validator) AddError(field, message string) {
v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

func (v *Validator) IsValid() bool {
return len(v.errors) == 0
}

func (v *Validator) ErrorMap() map[string]string {
result := make(map[string]string)
for _, err := range v.errors {
result[err.Field] = err.Message
}
return result
}

func ValidateIdentifyRequest(req models.IdentifyRequest) error {
v := New()

if req.Signals.Canvas2DHash == "" {
v.AddError("canvas_2d_hash", "required")
} else if req.Signals.Canvas2DHash != "error" && req.Signals.Canvas2DHash != "no_context" {
if !hashRegex.MatchString(req.Signals.Canvas2DHash) {
v.AddError("canvas_2d_hash", "invalid format")
}
}

if req.Signals.AudioHash == "" {
v.AddError("audio_hash", "required")
}

if req.Signals.HardwareConcurrency < 0 || req.Signals.HardwareConcurrency > 256 {
v.AddError("hardware_concurrency", "out of range")
}

if len(req.Signals.UserAgent) > 1000 {
v.AddError("user_agent", "too long")
}

if !v.IsValid() {
return fmt.Errorf("validation failed: %v", v.ErrorMap())
}
return nil
}

func SanitizeString(s string) string {
s = strings.ReplaceAll(s, "\x00", "")
var result strings.Builder
for _, r := range s {
if r >= 32 || r == '\n' || r == '\t' {
result.WriteRune(r)
}
}
return result.String()
}
