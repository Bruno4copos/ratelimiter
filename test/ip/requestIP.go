package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	url := "http://localhost:8080/hi"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("user-agent", "vscode-restclient")

	for i := 0; i < 200; i++ {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Erro na requisicão http: %v", err)
			panic(err)
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Erro no retorno da requisição http: %v", err)
			panic(err)
		}

		fmt.Println(res)
		fmt.Println(string(body))
	}
}
