package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/heroku/x/hmetrics/onload"
)

const (
	// ECounterPartyID – UUID существующего контрагента
	ECounterPartyID = "ff7e357d-20bc-11eb-0a80-03e30004b971"
	// ERetailStoreID – UUID существующей торговой точки
	ERetailStoreID = "87016576-f678-11e9-0a80-048c000a78b1"
)

// Counterparty struct
type Counterparty struct {
	ID                 string `json:"id"`                 // Уникальный идентификатор покупателя в системе лояльности в формате GUID Необходимое
	Name               string `json:"name"`               // ФИО покупателя Необходимое
	DiscountCardNumber string `json:"discountCardNumber"` // Номер дисконтной карты
	Phone              string `json:"phone"`              // Номер телефона в произвольном формате
	Email              string `json:"email"`              // Почтовый адрес
}

// DetailCounterParty struct
type DetailCounterParty struct {
	RetailStore struct {
		Meta struct {
			Href string `json:"href"` // Идентификатор точки продаж Необходимое
			ID   string `json:"id"`   // Идентификатор точки продаж Необходимое
		} `json:"meta"`
		Name string `json:"name"` // Название точки продаж
	} `json:"retailStore"`
	Meta struct {
		Href string `json:"href"` // Идентификатор покупателя Необходимое
		ID   string `json:"id"`   // Идентификатор покупателя Необходимое
	} `json:"meta"`
	Name               string `json:"name"`               // ФИО покупателя Необходимое
	DiscountCardNumber string `json:"discountCardNumber"` // Номер дисконтной карты
	Phone              string `json:"phone"`              // Номер телефона в произвольном формате
	Email              string `json:"email"`              // Почтовый адрес
}

// DetailCounterPartyResponse struct
type DetailCounterPartyResponse struct {
	BonusProgram `json:"bonusProgram"`
}

// BonusProgram struct
type BonusProgram struct {
	AgentBonusBalance int `json:"agentBonusBalance"`
}

// Counterparties struct
type Counterparties struct {
	Rows []Counterparty `json:"rows"`
}

// LognexAuthMiddleware checks headers token
func LognexAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("Lognex-Discount-API-Auth-Token"); token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// GetAgentBalance finds the bonus balance
func GetAgentBalance(retailStoreID, counterpartyID string) (bonusBalance int) {

	if retailStoreID == ERetailStoreID {
		if counterpartyID == ECounterPartyID {
			bonusBalance = 1500
		}
	} else {
		if counterpartyID == ECounterPartyID {
			bonusBalance = 50
		}
	}

	// default is 0
	return
}

// GetCounterParties finds the counterparty
func GetCounterParties(searchQuery string) *Counterparties {
	// search...
	return &Counterparties{
		Rows: []Counterparty{
			{ID: ECounterPartyID, Name: "TEST", DiscountCardNumber: "1111", Phone: "123456", Email: "a@b.c"},
		},
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(LognexAuthMiddleware)

	r.Post("/counterparty/detail", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var counterpartyDetailed DetailCounterParty
		err = json.Unmarshal(b, &counterpartyDetailed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bonusBalance := GetAgentBalance(counterpartyDetailed.RetailStore.Meta.ID, counterpartyDetailed.Meta.ID)

		detailResponse := &DetailCounterPartyResponse{
			BonusProgram{
				AgentBonusBalance: bonusBalance,
			},
		}

		output, err := json.Marshal(detailResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)
	})

	r.Get("/counterparty", func(w http.ResponseWriter, r *http.Request) {
		searchQuery := r.FormValue("search")

		response := GetCounterParties(searchQuery)
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	log.Printf("Server listen on port :%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
