package consts

var CANONICAL_FIELDS_AGGREGATION_MAP = map[string]string{
	"service":       "toString(attributes.service)",
	"level":         "lowerUTF8(toString(attributes.level))",
	"timestamp":     "toString(attributes.timestamp)",
	"environment":   "toString(attributes.environment)",
	"event":         "toString(attributes.event)",
	"error_message": "toString(attributes.error_message)",
	"status_code":   "toString(attributes.status_code)",
	"duration":      "toString(attributes.duration)",
	"ip":            "toString(attributes.ip)",
	"method":        "upperUTF8(toString(attributes.method))",
	"path":          "toString(attributes.path)",
	"useragent":     "toString(attributes.useragent)",
	"hostname":      "toString(attributes.hostname)",
}
