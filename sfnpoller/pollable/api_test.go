package pollable

import (
	"testing"
)

func TestTruncateErrorIfRequiredWhenStringIsGraterThan256Chars(t *testing.T) {
	errString := "Get \"https://upi.hdfcbank.com/oauth/token?client_id=client_id&amp;client_secret=client_secret&amp;grant_type=password&amp;password=password&amp;username=username\": read tcp: read: connection reset by peer read tcp: read: connection reset by peer read tcp: read: connection reset by peer"
	errStringTruncated := truncateErrorIfRequired(errString)

	if len(errStringTruncated) != 256 {
		t.Errorf("Truncated errString has %d character, expected 256", len(errString))
		return
	}

	expectedString := errString[:253] + "..."
	if expectedString != errStringTruncated {
		t.Errorf("Truncated errString expected %s, got %s", expectedString, errStringTruncated)
		return
	}
}

func TestTruncateErrorIfRequiredWhenErrorStringIs257Chars(t *testing.T) {
	errString := "R.SVWra,9KT&vh9;{/fmzG_85E!bMaFuVz*N8{[q;$_?kU3a7@c)6i@qt2dR@QZUJR00SPT/40[;{NP6Qdp@&JY{74,u([7V1D?P)G,u9VUbm[W$}#)uki8w2fN,r@*17hRjJu.?zhF)}-j(2AgRSbdiqna[},5CnX/-3zFYw-eR-+[0rY#QjgBmcb4#ywQMwNbNLjv+qd6(rHc+12}Q$T@t?7X)vxj{;+m.#]!XY8n:rJ+&!A&ejLkHT%/@QX%ii"
	errStringTruncated := truncateErrorIfRequired(errString)

	expectedString := errString[:253] + "..."
	if expectedString != errStringTruncated {
		t.Errorf("Truncated errString expected %s, got %s", errString, errStringTruncated)
		return
	}
}

func TestTruncateErrorIfRequiredWhenErrorStringIs256Chars(t *testing.T) {
	errString := "R.SVWra,9KT&vh9;{/fmzG_85E!bMaFuVz*N8{[q;$_?kU3a7@c)6i@qt2dR@QZUJR00SPT/40[;{NP6Qdp@&JY{74,u([7V1D?P)G,u9VUbm[W$}#)uki8w2fN,r@*17hRjJu.?zhF)}-j(2AgRSbdiqna[},5CnX/-3zFYw-eR-+[0rY#QjgBmcb4#ywQMwNbNLjv+qd6(rHc+12}Q$T@t?7X)vxj{;+m.#]!XY8n:rJ+&!A&ejLkHT%/@QX%i"
	errStringTruncated := truncateErrorIfRequired(errString)

	if errString != errStringTruncated {
		t.Errorf("Truncated errString expected %s, got %s", errString, errStringTruncated)
		return
	}
}

func TestTruncateErrorIfRequiredWhenErrorStringIsLessThan256Chars(t *testing.T) {
	errString := "read tcp 10.12.65.123:55362-&gt;175.100.163.28:443: read: connection reset by peer"
	errStringTruncated := truncateErrorIfRequired(errString)

	if errString != errStringTruncated {
		t.Errorf("Truncated errString expected %s, got %s", errString, errStringTruncated)
		return
	}
}
