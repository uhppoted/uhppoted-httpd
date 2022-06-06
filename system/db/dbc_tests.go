//go:build tests

package db

func DBCWithImpl(impl impl) DBC {
	return DBC{
		impl: impl,
	}
}
