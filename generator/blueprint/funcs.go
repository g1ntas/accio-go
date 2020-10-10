package blueprint

import (
	"github.com/lestrrat-go/strftime"
	"go.starlark.net/starlark"
	"time"
)

func predeclaredFuncs() starlark.StringDict {
	return starlark.StringDict{
		"strftime": starlark.NewBuiltin("strftime", builtinStrftime),
		"time":     starlark.NewBuiltin("time", builtinTime),
	}
}

func builtinStrftime(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var format string
	var timestamp int
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "format", &format, "time?", &timestamp); err != nil {
		return nil, err
	}
	var t time.Time
	if timestamp == 0 {
		t = time.Now().UTC()
	} else {
		t = time.Unix(int64(timestamp), 0).UTC()
	}
	out, err := strftime.Format(format, t)
	if err != nil {
		return nil, err
	}
	return starlark.String(out), nil
}

func builtinTime(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	timestamp := time.Now().UTC().Unix()
	return starlark.MakeInt64(timestamp), nil
}
