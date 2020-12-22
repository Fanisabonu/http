package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/Fanisabonu/http/cmd/app/middleware"
	"github.com/Fanisabonu/http/pkg/customers"
)

// Server представляет собой логический сервер нашего приложения.
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	// mw *middleware.Middleware
}

// NewServer ...
func NewServer(mux *mux.Router, customersSvc *customers.Service,) *Server {
	return &Server{mux: mux, customersSvc: customersSvc,}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

// MyStruct ...
type MyStruct struct {
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

// MySecondStruct ...
type MySecondStruct struct {
	Status     string `json:"status"`
	CustomerID int64  `json:"customerId"`
}

// Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	customersAuthenticateMd := middleware.Authenticate(s.customersSvc.IDByTokenForCustomers)

	customersSubrouter := s.mux.PathPrefix("/api/customers").Subrouter()
	customersSubrouter.Use(customersAuthenticateMd)
	customersSubrouter.HandleFunc("", s.handleCustomerRegistration).Methods("POST")
	customersSubrouter.HandleFunc("/token", s.handleCustomerGetToken).Methods("POST")
	customersSubrouter.HandleFunc("/token/validate", s.handleCustomerValidateToken).Methods("POST")
	customersSubrouter.HandleFunc("/products", s.handleCustomerGetProducts).Methods("GET")
	customersSubrouter.HandleFunc("/purchases", s.handleCustomerGetPurchases).Methods("GET")
	customersSubrouter.HandleFunc("/purchases", s.handleCustomerMakePurchase).Methods("POST")
	customersSubrouter.HandleFunc("/active", s.handleGetAllActiveCustomers).Methods("GET")
	customersSubrouter.HandleFunc("/{id}", s.handleGetCustomerByID).Methods("GET")
	customersSubrouter.HandleFunc("", s.handleGetAllCustomers).Methods("GET")
	customersSubrouter.HandleFunc("/{id}", s.handleRemoveCustomerByID).Methods("DELETE")
	customersSubrouter.HandleFunc("/{id}/block", s.handleBlockCustomerByID).Methods("POST")
	customersSubrouter.HandleFunc("/{id}/block", s.handleUnblockCustomerByID).Methods("DELETE")
	
	managerAuthenticateMd2 := middleware.Authenticate(s.customersSvc.IDByTokenForManagers2)
	managersSubrouter2 := s.mux.PathPrefix("/api/managers/sales").Subrouter()
	managersSubrouter2.Use(managerAuthenticateMd2)

	managersSubrouter3 := s.mux.PathPrefix("/api/managers/sales").Subrouter()
	managersSubrouter3.Use(managerAuthenticateMd2)

	managerAuthenticateMd := middleware.Authenticate(s.customersSvc.IDByTokenForManagers)
	managersSubrouter := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSubrouter.Use(managerAuthenticateMd)
	managersSubrouter.HandleFunc("", s.handleManagerRegistration).Methods("POST")
	managersSubrouter.HandleFunc("/token", s.handleManagerGetToken).Methods("POST")
	// managersSubrouter.HandleFunc("/token/validate", s.handleManagerValidateToken).Methods("POST")
	managersSubrouter3.HandleFunc("", s.handleManagerGetSales).Methods("GET")
	managersSubrouter2.HandleFunc("", s.handleManagerMakeSale).Methods("POST")
	s.mux.HandleFunc("/api/managers/products", s.handleManagerChangeProduct).Methods("POST")
	// managersSubrouter.HandleFunc("/customers", s.handleManagerGetCustomers).Methods("GET")
	// managersSubrouter.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods("POST")
	// managersSubrouter.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods("DELETE")
}


func (s *Server) handleManagerGetSales(writer http.ResponseWriter, request *http.Request)  {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if id == 0 {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, err := s.customersSvc.GetSales(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	total := &customers.GetSales{
		ManagerID: id,
		Total: result,
	}

	data, err := json.Marshal(total)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}


}


func (s *Server) handleManagerMakeSale(writer http.ResponseWriter, request *http.Request)  {

	valueID, err := middleware.Authentication(request.Context())  
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return 
	}

	if valueID == 0 {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var item *customers.MakeSale

	err = json.NewDecoder(request.Body).Decode(&item)
	
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	
	item.ManagerID = valueID

	sale, err := s.customersSvc.MakeSale(request.Context(), item)

	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(sale)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}


}



func (s *Server) handleManagerChangeProduct(writer http.ResponseWriter, request *http.Request)  {
	var item *customers.Product
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		return
	}

	result, err := s.customersSvc.SaveChangeProduct(request.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}


func (s *Server) handleManagerRegistration(writer http.ResponseWriter, request *http.Request)  {
	var item *customers.Manager
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		return
	}


	err = s.customersSvc.RegisterManager(request.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.customersSvc.TokenForManagerRegistr(request.Context(), item.Phone, item.Password)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}


	result := &customers.Token{
		Token: token,
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}



}


func (s *Server) handleManagerGetToken(writer http.ResponseWriter, request *http.Request)  {
	
	var item *customers.Auth

	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		return
	}

	token, err := s.customersSvc.TokenForManager(request.Context(), item.Phone, item.Password)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}
	
	result := &customers.Token{
		Token: token,
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}


func (s *Server) handleCustomerMakePurchase(writer http.ResponseWriter, request *http.Request)  {
	var item *customers.Purchase
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	purchase, err := s.customersSvc.MakePurchase(request.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(purchase)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}


func (s *Server) handleCustomerGetPurchases(writer http.ResponseWriter, request *http.Request)  {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items, err := s.customersSvc.Purchases(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleCustomerGetProducts(writer http.ResponseWriter, request *http.Request)  {
	items, err := s.customersSvc.Products(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}


func (s *Server) handleCustomerValidateToken(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	

	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	clientsID, err := s.customersSvc.AutenticateCustomer(request.Context(), item.Token)

	if err != nil {
		if err == customers.ErrNoSuchUser {
			result := &MyStruct{
				Status: "fail",
				Reason: "not found",
			}
			
			data, cerr := json.Marshal(result)
			if cerr != nil {
				log.Print(err)
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound)
			writer.Write(data)
			return

		} else if err == customers.ErrExpire {
			result := &MyStruct{
				Status: "fail",
				Reason: "expired",
			}
			data, cerr := json.Marshal(result)
			if cerr != nil {
				log.Print(err)
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(data)
			return
		}
	}

	result := &MySecondStruct{
		Status:     "ok",
		CustomerID: clientsID,
	}
	data, cerr := json.Marshal(result)
	if cerr != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}



func (s *Server) handleCustomerGetToken(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer

	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		return
	}

	token, err := s.customersSvc.TokenForCustomer(request.Context(), item.Login, item.Password)
	if err != nil {

		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}
	
	result := &customers.Customer{
		Token: token,
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}

}

func (s *Server) handleCustomerRegistration(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Registration
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	newCustomer, err := s.customersSvc.RegisterCustomer(request.Context(), item)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(newCustomer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetCustomerByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.customersSvc.ByID(request.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllCustomers(writer http.ResponseWriter, request *http.Request) {
	item, err := s.customersSvc.All(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllActiveCustomers(writer http.ResponseWriter, request *http.Request) {
	item, err := s.customersSvc.AllActive(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleRemoveCustomerByID(writer http.ResponseWriter, request *http.Request) {
	customerID, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	removedCustomer, err := s.customersSvc.RemoveByID(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(removedCustomer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleBlockCustomerByID(writer http.ResponseWriter, request *http.Request) {
	customerID, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	blockedUser, err := s.customersSvc.BlockUser(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(blockedUser)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleUnblockCustomerByID(writer http.ResponseWriter, request *http.Request) {
	customerID, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	unblockedUser, err := s.customersSvc.UnblockUser(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(unblockedUser)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}
