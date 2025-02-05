package endpoints

import (
	"fmt"
	"io"
	"net/http"
)

func GetTest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is test endpoint!\n")
}
