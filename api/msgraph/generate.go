package msgraph

//go:generate oapi-codegen -generate types -o types.go -package msgraph openapi.json
//DONT go:generate oapi-codegen -generate client -o client.go -package msgraph openapi.json

// Manual formatting fixes for the JSON if you update it:
//
// "format": "decimal" -> "format": "double"
// {:space:}NaN -> {:space:}"NaN FIXME DAMMIT"
