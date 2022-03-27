package gopwebtampering

import "unicode/utf8"

func maxSize(slice []string) int {
	max := 0

	for _, i := range slice {
		len := utf8.RuneCountInString(i)
		if len > max {
			max = len
		}
	}

	return max
}

var maxLenHeader = maxSize(HeadersIP)
var maxLenLocalhostAddresses = maxSize(LocalhostAddresses) + 3

var HeadersIP = []string{
	"X-Forwarded-For",
	"X-Real-IP",
	"X-ProxyUser-Ip",
	"Client-IP",
	"Forwarded-For-Ip",
	"Forwarded-For",
	"Forwarded-For",
	"Forwarded",
	"Forwarded",
	"True-Client-IP",
	"X-Client-IP",
	"X-Custom-IP-Authorization",
	"X-Forward-For",
	"X-Forward",
	"X-Forward",
	"X-Forwarded-By",
	"X-Forwarded-By",
	"X-Forwarded-For-Original",
	"X-Forwarded-For-Original",
	"X-Forwarded-For",
	"X-Forwarded-For",
	"X-Forwarded-Server",
	"X-Forwarded-Server",
	"X-Forwarded",
	"X-Forwarded",
	"X-Forwared-Host",
	"X-Forwared-Host",
	"X-Host",
	"X-Host",
	"X-HTTP-Host-Override",
	"X-Originating-IP",
	"X-Real-IP",
	"X-Remote-Addr",
	"X-Remote-Addr",
	"X-Remote-IP",
}

var LocalhostAddresses = []string{
	"127.0.0.1",
	"localhost",
	"0",
	"[::]",
	"[0000::1]",
	"[0:0:0:0:0:ffff:127.0.0.1]",
	"①②⑦.⓪.⓪.⓪",
	"127.127.127.127",
	"127.0.1.3",
	"127.0.0.0",
	"127。0。0。1",
	"127%E3%80%820%E3%80%820%E3%80%821",
	"192.168.0.1",
	"192.168.1.1",

	// 127.0.0.1
	"2130706433",

	// Octal Bypass
	"0177.0000.0000.0001",
	"00000177.00000000.00000000.00000001",
	"017700000001",

	// Hexadecimal bypass
	// 127.0.0.1 = 0x7f 00 00 01
	"0x7f000001",

	// DNS to localhost
	"spoofed.burpcollaborator.net",
}
