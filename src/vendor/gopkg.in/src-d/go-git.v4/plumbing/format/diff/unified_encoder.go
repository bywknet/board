	"errors"
var ErrBothFilesEmpty = errors.New("both files are empty")

		return ErrBothFilesEmpty
		c.current.AddOp(Equal, c.afterContext[:c.ctxLines]...)
		c.beforeContext = c.afterContext[c.ctxLines:]