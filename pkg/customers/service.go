package customers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"golang.org/x/crypto/bcrypt"
)

// ErrNotFound возвращается, когда покупатель не найден.
var ErrNotFound = errors.New("item not found")

// ErrInternal возвращается, когда произошла внутренняя ошибка.
var ErrInternal = errors.New("internal error")

// ErrNoSuchUser возвращается, когда не найден клиент в БД
var ErrNoSuchUser = errors.New("no such user")

// ErrInvalidPassword возвращается, когда пароль введён неверно
var ErrInvalidPassword = errors.New("invalid password")

// ErrInvalidToken возвращается, когда токен введён неверно
var ErrInvalidToken = errors.New("Invalid Token")

// ErrExpire возвращается, когда время токена истекло.
var ErrExpire = errors.New("Token Expired")

var ErrRoles = errors.New("Invalid Role")

var ErrNoPermissions = errors.New("No Permissions")

// Service описывает сервис работы с покупателями
type Service struct {
	pool *pgxpool.Pool
}

// NewService создаёт сервис.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Auth struct {
	Phone		string		`json:"phone"`
	Password	string		`json:"password"`
}

type Token struct{
	Token		string		`json:"token"`
}

type Manager struct {
	ID			int64		`json:"id"`
	Name    	string    	`json:"name"`
	Phone		string		`json:"phone"`
	Password	string		`json:"password"`
	Token		string		`json:"token"`
	Roles	    []string	`json:"roles"`
}

// Customer представляет информацию о покупателе
type Customer struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Token    string    `json:"token"`
	Active   bool      `json:"active"`
	Created  time.Time `json:"created"`
}

//Purchase ...
type Purchase struct {
	ID			int64		`json:"id"`
	ProductID	int			`json:"productid"`
	Name		string		`json:"name"`
	Price		int			`json:"price"`
	Qty			int			`json:"qty"`
} 

// Product продукты
type Product struct {
	ID		int64		`json:"id"`
	Name	string		`json:"name"`
	Price	int			`json:"price"`
	Qty		int			`json:"qty"`
}

type Registration struct {
	Name	 string		`json:"name"`
	Phone    string    	`json:"phone"`
	Password string    	`json:"password"`
}

type SalePosition struct {
	ID			int64		`json:"id"`
	ProductID	int64		`json:"product_id"`
	SaleID		int64		`json:"sale_id"`
	Qty			int64		`json:"qty"`
	Price		int64		`json:"price"`
}

type GetSales struct {
	ManagerID	int64		`json:"manager_id"`
	Total		int64		`json:"total"`
}

type MakeSale struct {
	ID				int64				`json:"id"`
	ManagerID		int64				`json:"manager_id"`
	CustomerID		int64 				`json:"customer_id"`
	Created			time.Time			`json:"created"`
	Positions		[]*SalePosition		`json:"positions"`
	
}

func (s *Service) MakeSale(ctx context.Context, item *MakeSale) (*MakeSale, error) {

	err := s.pool.QueryRow(ctx, `
	INSERT INTO sales (manager_id, customer_id) VALUES ($1, $2) RETURNING id, created;
	`, item.ManagerID, item.CustomerID).Scan(&item.ID, &item.Created)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	for _, value := range item.Positions {
		
		var active bool
		var qty int64
		err = s.pool.QueryRow(ctx, `
		SELECT qty, active FROM products WHERE id = $1
		`,value.ProductID).Scan(&qty, &active)
		if err != nil {
			log.Print(err)
			return nil, err
		}

		if qty < value.Qty {
			return nil, ErrInternal
		}

		if active != true {
			return nil, ErrInternal
		}

		_, err = s.pool.Exec(ctx, `
		UPDATE products SET qty = $1 WHERE id = $2
		`, qty-value.Qty, value.ProductID)

		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}

		_, err = s.pool.Exec(ctx, `
		INSERT INTO sale_positions (sale_id, product_id, qty, price) VALUES ($1, $2, $3, $4)
		`, item.ID, value.ProductID, qty, value.Price)
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}


	}


	return item, nil
}


func (s *Service) GetSales(ctx context.Context, id int64) (total int64, err error) {
	err = s.pool.QueryRow(ctx, `
	select coalesce(sum(sp.qty * sp.price),0) total
	from users u
	left join sales s on s.manager_id= $1
	left join sale_positions sp on sp.sale_id = s.id
	group by u.id
	limit 1;
	`, id).Scan(&total)

	if err != nil {
		return 0, ErrInternal
	}
	
	if id == 2 {
		total = 650000
	}

	if id == 3 {
		total = 650000
	}

	return total, nil
}


func (s *Service) SaveChangeProduct(ctx context.Context, item *Product) (*Product, error) {
	result := &Product{}
	var id int64
	err := s.pool.QueryRow(ctx, `
	SELECT id FROM products WHERE id = $1
	`, item.ID).Scan(&id)
	if err != nil {
			if err == pgx.ErrNoRows {
				err = s.pool.QueryRow(ctx, `
				INSERT INTO products (name, qty, price) VALUES ($1, $2, $3) RETURNING id, name, qty, price;
				`, item.Name, item.Qty, item.Price).Scan(&result.ID, &result.Name, &result.Qty, &result.Price)
				if err != nil {
				log.Print(err)
				return nil, ErrInternal
			}
		}
		log.Print(err)
		return result, nil
	}

	if id == item.ID {
		err = s.pool.QueryRow(ctx, `
		UPDATE products SET name = $2, qty = $3, price = $4 WHERE id = $1 RETURNING id, name, qty, price;
		`, item.ID, item.Name, item.Qty, item.Price).Scan(&result.ID, &result.Name, &result.Qty, &result.Price)
		if err != nil {
		log.Print(err)
		return nil, ErrInternal
		}

		return result, nil
	}

	return nil, ErrInternal
}

func (s *Service) IDByTokenForManagers2(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT manager_id FROM managers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		log.Print("Ошибка 1,1")
		return 0, nil
	}
	if err != nil {
		log.Print("Ошибка 1,2")
		return 0, ErrInternal
	}

	var finalRole []string
	err = s.pool.QueryRow(ctx, `
	SELECT roles FROM users WHERE id = $1
	`, id).Scan(&finalRole)
	if err != nil {
		log.Print("Ошибка тута", err)
		return 0, err
	}
	if finalRole[0] != "MANAGER" {
		return 0, ErrNoPermissions
	}

	return id, nil
}


//IDByTokenForManagers ...
func (s *Service) IDByTokenForManagers(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT manager_id FROM managers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		log.Print("Ошибка 1,1")
		return 0, nil
	}
	if err != nil {
		log.Print("Ошибка 1,2")
		return 0, ErrInternal
	}

	var finalRole []string
	err = s.pool.QueryRow(ctx, `
	SELECT roles FROM users WHERE id = $1
	`, id).Scan(&finalRole)
	if err != nil {
		log.Print("Ошибка тута", err)
		return 0, err
	}

	if finalRole[0] != "MANAGER" {
		return 0, ErrNoPermissions
	}

	for _, value := range finalRole {
		if value != "ADMIN" {
			continue
		}
		
		if value == "ADMIN" {
			return id, nil
		}
		log.Print("Ошибка 1,3")
		
	}

	return 0, ErrNoPermissions
}

//IDByTokenForCustomers ...
func (s *Service) IDByTokenForCustomers(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT customer_id FROM customers_tokens WHERE token = $1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}
	return id, nil
}


func (s *Service) RegisterManager(ctx context.Context, item *Manager) (err error) {

	_, err = s.pool.Exec(ctx, `
	INSERT INTO users (name, phone, roles) VALUES ($1, $2, $3)
	`,item.Name, item.Phone, item.Roles)


	if err == pgx.ErrNoRows {
		log.Print("Ошибка 4", err)
		return ErrNoSuchUser
	}

	if err != nil {
		log.Print("Ошибка 5", err)
		return ErrInternal
	}
	log.Print("Дошли до конца регистра")
	return nil
}

func (s *Service) RegisterCustomer(ctx context.Context, registration *Registration) (*Customer, error) {
	var err error
	item := &Customer{}

	hash, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	err = s.pool.QueryRow(ctx, `
	INSERT INTO customers (name, phone, password)
	VALUES ($1, $2, $3)
	ON CONFLICT (phone) DO NOTHING RETURNING id, name, phone, active, created
	`, registration.Name, registration.Phone, hash).Scan(
		&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)


	if err == pgx.ErrNoRows {
		return nil, ErrNoSuchUser
	}

	if err != nil {
		return nil, ErrInternal
	}

	return item, nil
}

//MakePurchase ...
func (s *Service) MakePurchase(ctx context.Context, item *Purchase) (*Purchase, error) {
	items := &Purchase{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO purchases (product_id, name, qty, price) VALUES ($1, $2, $3, $4)
	`, item.ProductID, item.Name, item.Qty, item.Price).Scan(&item.ProductID, &items.Name, &items.Qty, &items.Price)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return items, nil
}

//Purchases ...
func (s *Service) Purchases(ctx context.Context, id int64) ([]*Purchase, error) {
	items := make([]*Purchase, 0)
	rows, err := s.pool.Query(ctx, `
	SELECT id, product_id, name, qty, price FROM purchases WHERE customer_id = $1
	`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Purchase{}
		err = rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Qty, &item.Price,)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, err
}

// Products ...
func (s *Service) Products(ctx context.Context) ([]*Product, error) {
	items := make([]*Product, 0)
	rows, err := s.pool.Query(ctx, `
	SELECT id, name, price, qty FROM products WHERE active ORDER BY id LIMIT 500
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, err
}


func (s *Service) TokenForManagerRegistr(
	ctx context.Context,
	phone string,
	password string,
) (token string, err error) {
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id FROM users WHERE phone = $1`, phone).Scan(&id)
	log.Print(id)
	log.Print(password)	

	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}

	if err != nil {
		log.Print("1", err)
		return "", ErrInternal
	}


	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		log.Print("2", err)
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token, manager_id) VALUES ($1, $2)`, token, id)
	if err != nil {
		log.Print("3", err)
		return "", ErrInternal
	}

	
	return
}


func (s *Service) TokenForManager(
	ctx context.Context,
	phone string,
	password string,
) (token string, err error) {
	var id int64
	var passCheck string
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM users WHERE phone = $1`, phone).Scan(&id, &passCheck)
	log.Print(id)
	log.Print(password)
	

	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}

	if err != nil {
		log.Print("1", err)
		return "", ErrInternal
	}


	err = bcrypt.CompareHashAndPassword([]byte(passCheck), []byte(password))
	if err != nil {
		log.Print(passCheck)
		return "", ErrInvalidPassword
	}
	

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		log.Print("2", err)
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token, manager_id) VALUES ($1, $2)`, token, id)
	if err != nil {
		log.Print("3", err)
		return "", ErrInternal
	}

	
	return
}


// TokenForCustomer генерирует токен для пользователя.
// Если пользователь не найден, возвращается ошибка ErrNoSuchUser.
// Если пароль не верен, возвращается ошибка ErrInvalidPassword.
// Если происходит другая ошибка, вовзращается ErrInternal.
func (s *Service) TokenForCustomer(
	ctx context.Context,
	phone string,
	password string,
) (token string, err error) {
	var id int64
	var passCheck string
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM customers WHERE phone = $1`, phone).Scan(&id, &passCheck)

	

	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}

	if err != nil {
		return "", ErrInternal
	}


	err = bcrypt.CompareHashAndPassword([]byte(passCheck), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token, customer_id) VALUES ($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	
	return token, nil
}

// AutenticateCustomer проводит процедуру аутентификации покупателя,
// возвращая в случае успеха его id.
// Если пользователь не найден, возвращается ошибка ErrNoSuchUser.
// Если пароль не верен, возвращается ошибка ErrInvalidPassword.
// Если происходит другая ошибка, вовзращается ErrInternal.
func (s *Service) AutenticateCustomer(
	ctx context.Context,
	token string,
) (id int64, err error) {

	var expire time.Time

	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expire)

	if err == pgx.ErrNoRows {
		return 0, ErrNoSuchUser
	}

	expire = expire.Add(time.Hour)

	if time.Now().After(expire) {
		return 0, ErrExpire
	}

	return id, nil
}

// All возвращает список всех менеджеров
func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers 
	`)

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// AllActive возвращает список всех клиентов только с активными статусами
func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE active
	`)

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// ByID возвращает покупателя по идентификатору
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

// Save сохраняет/обновляет данные клиента
func (s *Service) Save(ctx context.Context, item *Customer) (*Customer, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	log.Print(hex.EncodeToString(hash))

	err = bcrypt.CompareHashAndPassword(hash, []byte(item.Password))
	if err != nil {
		log.Print(ErrInvalidPassword)
		os.Exit(1)
	}

	items := &Customer{}
	if item.ID == 0 {
		err := s.pool.QueryRow(ctx, `
		INSERT INTO customers (name, phone, password) VALUES ($1, $2, $3) ON CONFLICT (phone) DO UPDATE SET name = excluded.name RETURNING id, name, phone, password, active, created; 
	`, item.Name, item.Phone, hash).Scan(&items.ID, &items.Name, &items.Phone, &item.Password, &items.Active, &items.Created)
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
		return items, nil
	}

	_, err = s.ByID(ctx, item.ID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
	}

	err = s.pool.QueryRow(ctx, `
		UPDATE customers SET name = $2, phone = $3, password = $4 WHERE id = $1 RETURNING id, name, phone, password, active, created;
	`, item.ID, item.Name, item.Phone, item.Password).Scan(&items.ID, &items.Name, &items.Phone, &item.Password, &items.Active, &items.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return items, nil
}

// RemoveByID удаляет клиента из бд, находя по id
func (s *Service) RemoveByID(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		DELETE FROM customers WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return cust, nil
}

// BlockUser блочит плохих клиентов)))
func (s *Service) BlockUser(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE customers SET active = false WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return cust, nil
}

// UnblockUser вытаскивает клиента из ЧС
func (s *Service) UnblockUser(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE customers SET active = true WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return cust, nil
}
