[![Go](https://github.com/uponus/tnef/actions/workflows/go.yml/badge.svg)](https://github.com/uponus/tnef/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/uponus/tnef?status.svg)](https://godoc.org/github.com/uponus/tnef)
[![CodeQL](https://github.com/uponus/tnef/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/uponus/tnef/actions/workflows/codeql-analysis.yml)

With this library you can extract the body and attachments from Transport
Neutral Encapsulation Format (TNEF) files.

This work is based on https://github.com/koodaamo/tnefparse,
http://www.freeutils.net/source/jtnef/ and https://github.com/teamwork/tnef.

## Example usage

```go
package main
import (

	"io/ioutil"
	"os"
	"github.com/uponus/tnef"
)

func main() {
	t, err := tnef.DecodeFile("./winmail.dat")
	if err != nil {
		return
	}
	wd, _ := os.Getwd()
	for _, a := range t.Attachments {
		ioutil.WriteFile(wd+"/"+a.Title, a.Data, 0777)
	}
	ioutil.WriteFile(wd+"/bodyHTML.html", t.BodyHTML, 0777)
	ioutil.WriteFile(wd+"/bodyPlain.html", t.Body, 0777)
}
```
