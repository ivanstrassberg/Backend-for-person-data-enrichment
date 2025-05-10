package api

import (
	"context"
	db "db"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type APIServer struct {
	listenAddr string
	dbStorage  db.PostgresStorage
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiErr struct {
	Error string
}

type APIResponse struct {
	API      string
	Data     interface{}
	APIError string
}

func WriteJson(w http.ResponseWriter, code int, v any, logErr ...any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	log.Println("error", logErr)
	return json.NewEncoder(w).Encode(v)
}

type Router struct {
	mux *http.ServeMux
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, apiErr{Error: err.Error()})
		}
		incomingReq := fmt.Sprintf("%s %s", r.Method, strings.Split(r.Pattern, " ")[1])
		log.Printf("incoming request %s", incomingReq)
	}
}

func NewAPIServer(listenAddr string, postgresDB db.PostgresStorage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		dbStorage:  postgresDB,
	}
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func FetchAPI(ctx context.Context, apiURL string, param string, resultChan chan<- APIResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	fullURL := fmt.Sprintf("%s?name=%s", apiURL, param)

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		resultChan <- APIResponse{API: apiURL, APIError: err.Error()}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		resultChan <- APIResponse{API: apiURL, APIError: err.Error()}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resultChan <- APIResponse{API: apiURL, APIError: err.Error()}
		return
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		resultChan <- APIResponse{API: apiURL, APIError: err.Error()}
		return
	}

	resultChan <- APIResponse{API: apiURL, Data: data}
}

func ProcessExtAPIs(responses []APIResponse) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, resp := range responses {
		switch resp.API {
		case "https://api.agify.io/":
			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected result, map expected %T", resp.Data)
			}

			ageResp := AgeResp{
				Age: int(data["age"].(float64)),
			}
			result["age"] = ageResp.Age

		case "https://api.genderize.io/":
			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected result, map expected %T", resp.Data)
			}
			genderResp := GenderResp{
				Gender:      data["gender"].(string),
				Probability: data["probability"].(float64),
			}
			result["gender"] = genderResp.Gender
			result["gender_probability"] = genderResp.Probability

		case "https://api.nationalize.io/":
			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected result, map expected %T", resp.Data)
			}

			countries := make([]CountryRespMap, 0)
			for _, c := range data["country"].([]interface{}) {
				country := c.(map[string]interface{})
				countries = append(countries, CountryRespMap{
					CountryID:   country["country_id"].(string),
					Probability: country["probability"].(float64),
				})
			}

			nationalizeResp := NationalityResp{
				Country: countries,
			}

			if len(nationalizeResp.Country) > 0 {
				highest := nationalizeResp.Country[0]
				for _, c := range nationalizeResp.Country {
					if c.Probability > highest.Probability {
						highest = c
					}
				}
				result["country"] = highest.CountryID
				result["country_probability"] = highest.Probability
			}
		}
	}

	return result, nil
}

func FetchAPIS(name string) []APIResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resultChan := make(chan APIResponse, len(ExtAPIs))
	var wg sync.WaitGroup

	for _, api := range ExtAPIs {
		wg.Add(1)
		go FetchAPI(ctx, api, name, resultChan, &wg)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var responces []APIResponse
	for resp := range resultChan {
		responces = append(responces, resp)
	}
	return responces
}
