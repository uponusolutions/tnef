[![Go](https://github.com/uponus/tnef/actions/workflows/go.yml/badge.svg)](https://github.com/uponus/tnef/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/uponus/tnef?status.svg)](https://godoc.org/github.com/uponus/tnef)
[![CodeQL](https://github.com/uponus/tnef/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/uponus/tnef/actions/workflows/codeql-analysis.yml)

With this library you can extract the body and attachments from Transport
Neutral Encapsulation Format (TNEF) files.

This work is based on 
https://github.com/koodaamo/tnefparse,
http://www.freeutils.net/source/jtnef/,
https://github.com/teamwork/tnef and
https://github.com/verdammelt/tnef.


## Example usage

```go
package main

import (
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
		os.WriteFile(wd+"/"+a.Title, a.Data, 0777)
	}

	htmlBody, found := tnef.AttributeByMAPIName(t.MAPIAttributes, tnef.MAPIBodyHTML)
	if found {
		os.WriteFile(wd+"/bodyHTML.html", htmlBody.Data, 0777)
	}

	txtBody, found := tnef.AttributeByMAPIName(t.MAPIAttributes, tnef.MAPIBody)
	if found {
		os.WriteFile(wd+"/bodyTxt.html", txtBody.Data, 0777)
	}
}
```
