// Copyright 2016 Alex Fluter

package bot

import (
	"crypto/x509"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

func IsChannel(name string) bool {
	return name[0] == '#'
}

func IsNick(name string) bool {
	return true
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func numeric(cmd string) bool {
	if len(cmd) != 3 {
		return false
	}
	if !isDigit(cmd[0]) {
		return false
	}
	if !isDigit(cmd[1]) {
		return false
	}
	if !isDigit(cmd[2]) {
		return false
	}
	return true
}

func sp(single, plural string, num int) string {
	if num == 1 {
		return single
	}
	return plural
}

func unixTimeStr(ss string) string {
	const TimeFmt = "Mon, 02 Jan 2006 15:04:05"
	var sec int
	var unix time.Time

	sec, _ = strconv.Atoi(ss)
	unix = time.Unix(int64(sec), 0)

	return unix.Format(TimeFmt)
}

// TODO: cert
func dumpCert(cert *x509.Certificate) string {
	return fmt.Sprintf("subject `%s', issue `%s'",
		cert.Subject,
		cert.Issuer)
}

func dateTime(t time.Time) string {
	y, m, d := t.Date()
	h, mi, s := t.Clock()
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		y, m, d, h, mi, s)
}

func bt() {
	debug.PrintStack()
}
