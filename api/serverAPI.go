package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var ExtAPIs []string = []string{
	"https://api.nationalize.io/",
	"https://api.genderize.io/",
	"https://api.agify.io/",
}

func (s *APIServer) RunAPIServer() {
	router := NewRouter()
	router.HandleEndpoints(s)
	log.Printf("server started on %s \n", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, router.mux); err != nil {
		fmt.Errorf("failed to start a server on %s", s.listenAddr)
	}
}

func (r Router) HandleEndpoints(s *APIServer) {
	m := r.mux
	m.HandleFunc("GET /people", makeHTTPHandleFunc(s.handleGetPeople))
	m.HandleFunc("POST /people", makeHTTPHandleFunc(s.handleCreatePeople))
	m.HandleFunc("PUT /people/{id}", makeHTTPHandleFunc(s.handleUpdatePeopleSkipEnrich))
	m.HandleFunc("PUT /people/enrich/{id}", makeHTTPHandleFunc(s.handleUpdatePeopleEnrich))
	m.HandleFunc("DELETE /people/{id}", makeHTTPHandleFunc(s.handleDeletePeople))
}

func (s *APIServer) handleGetPeople(w http.ResponseWriter, r *http.Request) error {
	return WriteJson(w, http.StatusOK, "ok")
}

func (s *APIServer) handleCreatePeople(w http.ResponseWriter, r *http.Request) error {
	person := new(PersonReq)
	if err := json.NewDecoder(r.Body).Decode(person); err != nil {
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	results := FetchAPIS(person.Name)

	process, err := ProcessExtAPIs(results)
	if err != nil {
		fmt.Println("err", err)
	}
	var enrichedPerson PersonEnriched = PersonEnriched{
		PersonReq:   *person,
		Age:         process["age"].(int),
		Nationality: process["country"].(string),
		Gender:      process["gender"].(string),
	}

	if err := s.dbStorage.CreatePerson(enrichedPerson.PersonReq.Name, enrichedPerson.PersonReq.Surname, enrichedPerson.PersonReq.Patronymic, enrichedPerson.Age, enrichedPerson.Gender, enrichedPerson.Nationality); err != nil {
		log.Printf("err: %s", err)
		WriteJson(w, http.StatusInternalServerError, "internal server error")
		return nil
	}

	return WriteJson(w, http.StatusOK, "ok")
}

func (s *APIServer) handleUpdatePeopleSkipEnrich(w http.ResponseWriter, r *http.Request) error {
	// idStr := r.PathValue("id")
	// id, err := strconv.Atoi(idStr)

	// if err != nil {
	// 	log.Printf("err: %s", err)
	// 	WriteJson(w, http.StatusBadRequest, "bad request")
	// 	return nil
	// }

	return WriteJson(w, http.StatusOK, "ok")
}
func (s *APIServer) handleUpdatePeopleEnrich(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Printf("err at create: %s", err)
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	person := new(PersonReq)
	if err := json.NewDecoder(r.Body).Decode(person); err != nil {
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	currName, err := s.dbStorage.CheckName(id)
	if err != nil {
		log.Printf("err at check: %s", err)
		WriteJson(w, http.StatusInternalServerError, "internal server error")
		return nil
	}
	if person.Name != "" && person.Name != currName {
		results := FetchAPIS(person.Name)

		processed, err := ProcessExtAPIs(results)

		if err != nil {
			return fmt.Errorf("failed to enrich data: %w", err)
		}

		var enrichedPerson PersonEnriched = PersonEnriched{
			PersonReq:   *person,
			Age:         processed["age"].(int),
			Nationality: processed["country"].(string),
			Gender:      processed["gender"].(string),
		}
		if err := s.dbStorage.UpdatePerson(id, enrichedPerson.PersonReq.Name, enrichedPerson.PersonReq.Surname, enrichedPerson.PersonReq.Patronymic, enrichedPerson.Age, enrichedPerson.Gender, enrichedPerson.Nationality); err != nil {
			log.Printf("err at update: %s", err)

			WriteJson(w, http.StatusNotFound, "internal server error")
			return nil
		}
	}

	return WriteJson(w, http.StatusOK, "ok")
}

func (s *APIServer) handleDeletePeople(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("err at delete: %s", err)
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	if err := s.dbStorage.DeletePerson(id); err != nil {
		log.Printf("err: %s", err)
		WriteJson(w, http.StatusNotFound, "internal server error")
		return nil

	}
	return WriteJson(w, http.StatusOK, "ok")
}
