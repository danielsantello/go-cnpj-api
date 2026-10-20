package company

import "errors"

const MaximumSearchResults = 1000

var ErrSearchTooBroad = errors.New("too many results; refine your search")
