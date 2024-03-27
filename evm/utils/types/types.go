package types

import "strings"

type Storage struct {
	AstId    int    `json:"astId"`
	Contract string `json:"contract"`
	Label    string `json:"label"`
	Offset   int    `json:"offset"`
	Slot     string `json:"slot"`
	Type     string `json:"type"`
}

type StorageKeyType struct {
	Encoding      string    `json:"encoding"`
	Key           string    `json:"key"`
	Label         string    `json:"label"`
	NumberOfBytes string    `json:"numberOfBytes"`
	Value         string    `json:"value"`
	Members       []Storage `json:"members"`
}

type ContractStorage struct {
	Storage []Storage                 `json:"storage"`
	Types   map[string]StorageKeyType `json:"types"`
}

type ContractVariable struct {
	Name  string `json:"variableName"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func IsDynamicType(typeName string) bool {
	return strings.HasPrefix(typeName, "t_bool") || strings.HasPrefix(typeName, "t_enum") || strings.HasPrefix(typeName, "t_uint") || strings.HasPrefix(typeName, "t_int") ||
		strings.HasPrefix(typeName, "t_address") || strings.HasPrefix(typeName, "t_contract") || strings.HasPrefix(typeName, "t_string") || strings.HasPrefix(typeName, "t_bytes")
}
