package echo

import (
	"github.com/HongJaison/go-admin3/tests/common"
	"github.com/gavv/httpexpect"
	"net/http"
	"testing"
)

func TestEcho(t *testing.T) {
	common.ExtraTest(httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(newHandler()),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
	}))
}
