package api

import (
	"db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	httpSwagger "github.com/swaggo/http-swagger"
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
	m.Handle("/swagger/", httpSwagger.WrapHandler)

	m.HandleFunc("GET /people", makeHTTPHandleFunc(s.handleGetPeopleWithPagination))

	m.HandleFunc("POST /people", makeHTTPHandleFunc(s.handleCreatePeople))

	m.HandleFunc("PATCH /people/{id}", makeHTTPHandleFunc(s.handleUpdatePeopleSkipEnrich))

	m.HandleFunc("PUT /people/enrich/{id}", makeHTTPHandleFunc(s.handleUpdatePeopleEnrich))

	m.HandleFunc("DELETE /people/{id}", makeHTTPHandleFunc(s.handleDeletePeople))
}

type PaginatedFilteredResults struct {
	Page           int         `json:"page"`
	PagesTotal     int         `json:"pages_total"`
	EntriesTotal   int         `json:"entries_total"`
	EntriesPerPage int         `json:"entries_per_page"`
	People         []db.Person `json:"people"`
}

// @Summary Список людей с фильтрацией и пагинацией
// @Description Получить пагинированный список людей с возможностью фильтрации по различным параметрам
// @Tags people
// @Accept  json
// @Produce  json
// @Param fname query string false "Фильтрация по имени (частичное совпадение)"
// @Param surname query string false "Фильтрация по фамилии (частичное совпадение)"
// @Param patronymic query string false "Фильтрация по отчество"
// @Param age query int false "Фильтрация по возрасту"
// @Param nationality query string false "Фильтрация по национальности"
// @Param gender query string false "Фильтрация по полу"
// @Param page query int false "Номер страницы (по умолчанию: 1)" default(1)
// @Param entries query int false "Количество записей на странице (по умолчанию: 10)" default(10)
// @Success 200 {object} PaginatedFilteredResults
// @Failure 400 {object} ApiError
// @Failure 500 {object} ApiError
// @Router /people [get]
func (s *APIServer) handleGetPeopleWithPagination(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()

	fname := query.Get("fname")
	surname := query.Get("surname")
	patronymic := query.Get("patronymic")
	ageStr := query.Get("age")
	nationality := query.Get("nationality")
	gender := query.Get("gender")

	page := parseIntPagination(query.Get("page"), 1)
	entries := parseIntPagination(query.Get("entries"), 10)
	offset := (page - 1) * entries

	var age int
	if ageStr != "" {
		parsedAge, err := strconv.Atoi(ageStr)
		if err != nil || parsedAge < 1 {
			WriteJson(w, http.StatusBadRequest, "invalid age")
			return err
		}
		age = parsedAge
	}

	people, total, err := s.dbStorage.GetPeopleWithPagination(fname, surname, patronymic, age, nationality, gender, entries, offset)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "internal server error")
		return err
	}

	pagesTotal := (total + entries - 1) / entries
	response := PaginatedFilteredResults{
		Page:           page,
		PagesTotal:     pagesTotal,
		EntriesTotal:   total,
		EntriesPerPage: entries,
		People:         people,
	}

	WriteJson(w, http.StatusOK, response)
	return nil
}

func parseIntPagination(s string, count int) int {
	countInt, err := strconv.Atoi(s)
	if err != nil || countInt < 1 {
		return count
	}
	return countInt
}

// @Summary Создание нового человека с обогащением данных
// @Description Создание новой записи о человеке с автоматическим обогащением данных из внешних API
// @Tags people
// @Accept  json
// @Produce  json
// @Param person body PersonReq true "Данные о человеке"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ApiError
// @Failure 500 {object} ApiError
// @Router /people [post]
func (s *APIServer) handleCreatePeople(w http.ResponseWriter, r *http.Request) error {
	person := new(PersonReq)
	if err := json.NewDecoder(r.Body).Decode(person); err != nil {
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	results := FetchAPIS(person.Name)

	process, err := ProcessExtAPIs(results)
	if err != nil {
		fmt.Println("err apis", err)
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

// @Summary Обновление данных человека без обогащения
// @Description Частичное обновление записи о человеке (без обогащения данных)
// @Tags people
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Param person body PersonEnriched true "Частичные данные человека"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ApiError
// @Failure 404 {object} ApiError
// @Failure 500 {object} ApiError
// @Router /people/{id} [patch]
func (s *APIServer) handleUpdatePeopleSkipEnrich(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Printf("err at fmt: %s", err)
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	enrichedPerson := new(PersonEnriched)
	if err := json.NewDecoder(r.Body).Decode(enrichedPerson); err != nil {
		WriteJson(w, http.StatusBadRequest, "bad request")
		return nil
	}

	if err := s.dbStorage.UpdatePersonPatch(id, enrichedPerson.PersonReq.Name, enrichedPerson.PersonReq.Surname, enrichedPerson.PersonReq.Patronymic, enrichedPerson.Age, enrichedPerson.Gender, enrichedPerson.Nationality); err != nil {
		log.Printf("err at update: %s", err)

		WriteJson(w, http.StatusNotFound, "internal server error")
		return nil
	}

	return WriteJson(w, http.StatusOK, "ok")
}

// @Summary Обновление данных человека с обогащением
// @Description Обновление записи о человеке с возможным обогащением данных в случае изменения имени
// @Tags people
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Param person body PersonReq true "Данные о человеке"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ApiError
// @Failure 404 {object} ApiError
// @Failure 500 {object} ApiError
// @Router /people/enrich/{id} [put]
func (s *APIServer) handleUpdatePeopleEnrich(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Printf("err fmt: %s", err)
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
		if err := s.dbStorage.UpdatePersonEnrich(id, enrichedPerson.PersonReq.Name, enrichedPerson.PersonReq.Surname, enrichedPerson.PersonReq.Patronymic, enrichedPerson.Age, enrichedPerson.Gender, enrichedPerson.Nationality); err != nil {
			log.Printf("err at update: %s", err)

			WriteJson(w, http.StatusNotFound, "internal server error")
			return nil
		}
	}

	return WriteJson(w, http.StatusOK, "ok")
}

// @Summary Удаление человека
// @Description Удаление записи о человеке по ID
// @Tags people
// @Accept  json
// @Produce  json
// @Param id path int true "ID человека"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ApiError
// @Failure 404 {object} ApiError
// @Failure 500 {object} ApiError
// @Router /people/{id} [delete]
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

type ApiError struct {
	Error string `json:"error" example:"error message"`
}

type SuccessResponse struct {
	Status string `json:"status" example:"success"`
}
