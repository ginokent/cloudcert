package errors

import "golang.org/x/xerrors"

func Errorf(format string, a ...interface{}) error {
	//nolint: wrapcheck
	return xerrors.Errorf(format, a...)
}
