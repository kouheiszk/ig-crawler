package crawler

import (
	"fmt"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const ErrorDelay = 30 * time.Second
const RequestTimeout = 30 * time.Second

func fetch(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return fetchWithRequest(request)
}

func fetchWithRequest(request *http.Request) ([]byte, error) {
	command, _ := http2curl.GetCurlCommand(request)
	log.Println(command)

	client := &http.Client{
		Timeout: RequestTimeout,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Print(errors.Wrap(err, "connection issue:"))
		time.Sleep(ErrorDelay)
		return fetchWithRequest(request)
	}
	defer response.Body.Close()

	if response.StatusCode == 429 {
		log.Printf("throtteling \"%s\"", request.URL)
		time.Sleep(ErrorDelay)
		return fetchWithRequest(request)
	}

	if response.StatusCode == 404 {
		log.Printf("not found \"%s\"", request.URL)
		return nil, fmt.Errorf("not found \"%s\"", request.URL)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("unable to read the response body")
	}

	return bytes, nil
}
