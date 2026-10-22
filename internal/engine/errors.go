package engine

import (
	"errors"
	"fmt"
	"strings"
)

type renderFileError struct {
	chain []string
	err   error
}

func wrapRenderFileError(name string, err error) error {
	if err == nil || name == "" {
		return err
	}

	var fileErr *renderFileError
	if errors.As(err, &fileErr) {
		chain := make([]string, 0, len(fileErr.chain)+1)
		chain = append(chain, name)
		chain = append(chain, fileErr.chain...)
		return &renderFileError{
			chain: chain,
			err:   fileErr.err,
		}
	}

	return &renderFileError{
		chain: []string{name},
		err:   err,
	}
}

func (e *renderFileError) Error() string {
	return fmt.Sprintf("模板渲染失败 [%s]: %v", strings.Join(e.chain, " -> "), e.err)
}

func (e *renderFileError) Unwrap() error {
	return e.err
}
