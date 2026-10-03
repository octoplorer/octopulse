package server

import (
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io"
	"maps"

	"github.com/danielgtaylor/huma/v2"
)

func init() {
	// Match array schemas to the API's JSON encoding policy. Set this once,
	// before registration, so concurrent server creation never writes the global.
	huma.DefaultArrayNullable = false
}

func apiJSONFormats(formats map[string]huma.Format) map[string]huma.Format {
	formats = maps.Clone(formats)
	// Huma uses "json" as the fallback for structured +json content types,
	// including application/problem+json error responses.
	for _, contentType := range []string{"application/json", "json"} {
		format := formats[contentType]
		format.Marshal = func(w io.Writer, v any) error {
			return jsonv2.MarshalWrite(w, v,
				json.DefaultOptionsV1(),
				jsonv2.FormatNilSliceAsNull(false),
				jsontext.EscapeForHTML(false),
			)
		}
		formats[contentType] = format
	}
	return formats
}
