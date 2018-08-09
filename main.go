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
 * select`get_etld_p1`('http://a.very.complex-domain.co.uk:8080/foo/bar'); -- outputs 'complex-domain.co.uk'
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
	"net"
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

	if a[0] == "" {
		return nil
	}

	// Replace back slashes with forward slashes since they aren't
	// really allowed but not impossible and we want to be forgiving
	r := strings.Replace(strings.ToLower(strings.TrimSpace(a[0])), `\`, `/`, -1)

	// Parsing first to get the hostname by itself, since `EffectiveTLDPlusOne`
	// acts strangely if it isn't just given the hostname
	u, err := url.Parse(r)
	if err != nil {
		l.Println(err.Error())
		return nil
	}

	// if the scheme isn't there we don't get what we expect (like a blank string
	// for 'com.s3-website-us-east-1.amazonaws.com')
	if u.Scheme == "" {
		u, err = url.Parse("http://" + r)
		if err != nil {
			l.Println(err.Error())
			return nil
		}
	}

	hostname := u.Hostname()

	// If the scheme isn't one of these, then it's likely
	// an Android app or something similar, in which case
	// TLD searching won't make sense here, and we'd expect the full address
	// not including the scheme
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ftp" {
		*length = uint64(len(hostname))
		return C.CString(hostname)
	}

	// localhost doesn't have a tld
	// if strings.IndexByte(hostname, '.') == -1 {
	if hostname == "localhost" {
		*length = uint64(len(hostname))
		return C.CString(hostname)
	}

	// Check to see if the hostname is an IP address since
	// they don't have TLDs
	addr := net.ParseIP(hostname)
	if addr != nil {
		*length = uint64(len(hostname))
		return C.CString(hostname)
	}

	h, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil || len(h) == 0 {
		if err != nil {
			l.Println(err.Error())
		} else {
			l.Println("Domain return string is blank")
		}
		return nil
	}

	*length = uint64(len(h))
	return C.CString(h)
}

func main() {}
