package msgraphbeta

//go:generate oapi-codegen -generate types -o types.go -package msgraphbeta openapi.json
//DONT go:generate oapi-codegen -generate client -o client.go -package msgraphbeta openapi.json

// Manual formatting fixes for the JSON if you update it:
//
// "format": "decimal" -> "format": "double"
// {:space:}NaN -> {:space:}"NaN FIXME DAMMIT"
