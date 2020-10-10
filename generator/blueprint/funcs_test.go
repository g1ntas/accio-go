package blueprint

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"os"
	"strconv"
	"testing"
	"time"
)

var mockCtx, _ = newContext(map[string]interface{}{})

func TestStrftime(t *testing.T) {
	script := wrapInlineScript(`strftime("%j")`)

	val, err := execute(script, &mockCtx)
	require.NoError(t, err)

	require.IsType(t, starlark.String(""), val)
	require.Equal(t, strconv.Itoa(time.Now().YearDay()), val.(starlark.String).GoString())
}

func TestStrftimeWithTimestamp(t *testing.T) {
	script := wrapInlineScript(`strftime("%A %a %B %b %C %c %D %d %e %F %H %h %I %j %k %l %M %m %n %p %R %r %S %T %t %U %u %V %v %W %w %X %x %Y %y %Z %z", 1136239445)`)

	os.Setenv("LC_ALL", "C")

	val, err := execute(script, &mockCtx)
	require.NoError(t, err)

	require.IsType(t, starlark.String(""), val)
	require.Equal(t, "Monday Mon January Jan 20 Mon Jan  2 22:04:05 2006 01/02/06 02  2 2006-01-02 22 Jan 10 002 22 10 04 01 \n PM 22:04 10:04:05 PM 05 22:04:05 \t 01 1 01  2-Jan-2006 01 1 22:04:05 01/02/06 2006 06 UTC +0000", val.(starlark.String).GoString())
}

func TestTime(t *testing.T) {
	script := wrapInlineScript(`time()`)

	realTime := time.Now().Unix()
	val, err := execute(script, &mockCtx)
	require.NoError(t, err)

	require.IsType(t, starlark.Int{}, val)
	evalTime, ok := val.(starlark.Int).Int64()
	require.True(t, ok, "Evaluated value is not a valid integer")

	require.Equal(t, evalTime, realTime)

}
