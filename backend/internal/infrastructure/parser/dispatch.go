package parser

import "strings"

// NewParserForPath returns the correct parser based on file extension.
// Files ending in ".unm" (but not ".unm.yaml") use DSLParser.
// Everything else uses YAMLParser.
func NewParserForPath(path string) Parser {
	if strings.HasSuffix(path, ".unm") && !strings.HasSuffix(path, ".unm.yaml") {
		return NewDSLParser()
	}
	return NewYAMLParser()
}
