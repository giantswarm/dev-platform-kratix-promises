package validation

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaValidator handles OpenAPI v3 schema validation using a proper JSON schema library
type SchemaValidator struct {
	logger   *slog.Logger
	compiler *jsonschema.Compiler
}

// NewSchemaValidator creates a new SchemaValidator with proper JSON schema support
func NewSchemaValidator() *SchemaValidator {
	compiler := jsonschema.NewCompiler()

	return &SchemaValidator{
		logger:   slog.Default().With("component", "schema-validator"),
		compiler: compiler,
	}
}

// ValidateSpec validates a resource specification against an OpenAPI v3 schema
func (v *SchemaValidator) ValidateSpec(schema map[string]interface{}, spec map[string]interface{}) (*resources.ValidationDetails, error) {
	v.logger.Debug("Starting JSON schema validation")

	details := &resources.ValidationDetails{
		Errors:   []resources.ValidationError{},
		Warnings: []resources.ValidationError{},
	}

	// Add the schema as a resource to the compiler
	schemaURL := "http://example.com/schema.json"
	err := v.compiler.AddResource(schemaURL, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to add schema to compiler: %w", err)
	}

	// Compile the schema
	compiledSchema, err := v.compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	// Validate the spec against the compiled schema
	err = compiledSchema.Validate(spec)
	if err != nil {
		// Handle validation errors from the JSON schema library
		if validationErr, ok := err.(*jsonschema.ValidationError); ok {
			v.convertJSONSchemaErrors(validationErr, details)
		} else {
			// Handle other types of errors
			details.Errors = append(details.Errors, resources.ValidationError{
				Field:   "",
				Message: fmt.Sprintf("Validation error: %s", err.Error()),
				Value:   spec,
			})
		}
	}

	v.logger.Debug("JSON schema validation completed", "errors", len(details.Errors), "warnings", len(details.Warnings))
	return details, nil
}

// convertJSONSchemaErrors converts errors from the JSON schema library to our internal format
func (v *SchemaValidator) convertJSONSchemaErrors(validationErr *jsonschema.ValidationError, details *resources.ValidationDetails) {
	// Convert the primary error
	fieldPath := "/" + strings.Join(validationErr.InstanceLocation, "/")
	details.Errors = append(details.Errors, resources.ValidationError{
		Field:   fieldPath,
		Message: validationErr.Error(),
		Value:   nil, // The library doesn't expose the actual value easily
	})

	// Convert any nested errors (for complex validations like anyOf, allOf, etc.)
	for _, cause := range validationErr.Causes {
		if cause != nil {
			nestedFieldPath := "/" + strings.Join(cause.InstanceLocation, "/")
			details.Errors = append(details.Errors, resources.ValidationError{
				Field:   nestedFieldPath,
				Message: cause.Error(),
				Value:   nil,
			})
		}
	}
}
