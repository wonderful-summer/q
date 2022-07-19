package q

import (
	"fmt"
	qt "github.com/wonderful-summer/q-type"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(message string) string {
	var pcs [32]uintptr
	// todo: 看文档这是什么意思
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback: ")

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() qt.HandlerFunc {
	return func(c qt.DefaultContext) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Status(http.StatusInternalServerError).End("Internal Server Error")
			}
		}()

		c.Next()
	}
}
