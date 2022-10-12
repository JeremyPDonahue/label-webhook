package main

const (
	addOperation     = "add"
	removeOperation  = "remove"
	replaceOperation = "replace"
	copyOperation    = "copy"
	moveOperation    = "move"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from"`
	Value interface{} `json:"value,omitempty"`
}

func addPatchOperation(path string, value interface{}) patchOperation {
	return patchOperation{
		Op:    addOperation,
		Path:  path,
		Value: value,
	}
}

// RemovePatchOperation returns a remove JSON patch operation.
func removePatchOperation(path string) patchOperation {
	return patchOperation{
		Op:   removeOperation,
		Path: path,
	}
}

// ReplacePatchOperation returns a replace JSON patch operation.
func replacePatchOperation(path string, value interface{}) patchOperation {
	return patchOperation{
		Op:    replaceOperation,
		Path:  path,
		Value: value,
	}
}

// CopyPatchOperation returns a copy JSON patch operation.
func copyPatchOperation(from, path string) patchOperation {
	return patchOperation{
		Op:   copyOperation,
		Path: path,
		From: from,
	}
}

// MovePatchOperation returns a move JSON patch operation.
func movePatchOperation(from, path string) patchOperation {
	return patchOperation{
		Op:   moveOperation,
		Path: path,
		From: from,
	}
}
