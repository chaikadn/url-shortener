package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func main() {
	endpoint := "http://localhost:8080/"

	fmt.Print("Введите длинный URL: ")
	reader := bufio.NewReader(os.Stdin)
	longURL, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	longURL = strings.TrimSpace(longURL)

	client := resty.New()

	response, err := client.R().
		SetBody(longURL).
		Post(endpoint)
	if err != nil {
		panic(err)
	}

	fmt.Println("Статус-код:", response.Status())
	fmt.Println(string(response.Body()))
}
