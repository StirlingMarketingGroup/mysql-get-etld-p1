/*
 * USAGE INSTRUCTIONS:
 *
 * make sure libmysqlclient-dev is installed:
 * apt install libmysqlclient-dev
 *
 * Replace "/usr/lib/mysql/plugin" with your MySQL plugins directory (can be found by running "select @@plugin_dir;")
 * go build -buildmode=c-shared -o get_etld_p1.so && cp get_etld_p1.so /usr/lib/mysql/plugin/get_etld_p1.so
 *
 * Then, on the server:
 * create function`get_etld_p1`returns string soname'get_etld_p1.so';
 *
 * And use/test like:
 * select`get_etld_p1`('https://stackoverflow.com/questions/51446087/how-to-debug-dump-go-variable-while-building-with-cgo?noredirect=1#comment89863750_51446087', 'noredirect'); -- outputs 'stackoverflow.com'
 *
 * Yeet!
 * Brian Leishman
 *
 */

package main

// #include <stdio.h>
// #include <sys/types.h>
// #include <sys/stat.h>
// #include <stdlib.h>
// #include <string.h>
// #include <mysql.h>
// #cgo CFLAGS: -O3 -I/usr/include/mysql -fno-omit-frame-pointer
import "C"
import (
	"log"
	"net/url"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/net/publicsuffix"
)

func msg(message *C.char, s string) {
	m := C.CString(s)
	defer C.free(unsafe.Pointer(m))

	C.strcpy(message, m)
}

//export get_etld_p1_deinit
func get_etld_p1_deinit(initid *C.UDF_INIT) {}

//export get_etld_p1_init
func get_etld_p1_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count != 1 {
		msg(message, "`get_etld_p1` requires 1 parameter: the URL string")
		return 1
	}

	argsTypes := (*[2]uint32)(unsafe.Pointer(args.arg_type))

	argsTypes[0] = C.STRING_RESULT
	initid.maybe_null = 1

	return 0
}

//export get_etld_p1
func get_etld_p1(initid *C.UDF_INIT, args *C.UDF_ARGS, result *C.char, length *uint64, isNull *C.char, message *C.char) *C.char {
	l := log.New(os.Stderr, "", 1)

	c := 1
	argsArgs := (*[1 << 30]*C.char)(unsafe.Pointer(args.args))[:c:c]

	a := make([]string, c, c)
	for i, argsArg := range argsArgs {
		// This should be the correct way, but lengths come through as "0"
		// for everything after the first argument, so hopefully we don't
		// encounter any URLs or param names with null bytes in them (not really that worried)
		// a[i] = C.GoStringN(argsArg, C.int(argsLengths[i]))

		a[i] = C.GoString(argsArg)
	}

	u, err := url.Parse(a[0])
	if err != nil {
		l.Println(err.Error())
		return nil
	}

	h, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil || len(h) == 0 {
		if err != nil {
			l.Println(err.Error())
		} else {
			l.Println("Domain return string is blank")
		}
		return nil
	}

	h = strings.ToLower(h)

	*length = uint64(len(h))
	return C.CString(h)
}

func main() {}
