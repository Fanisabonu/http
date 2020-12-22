package security

// import (
// 	"context"
// 	"log"

// 	"github.com/jackc/pgx/v4/pgxpool"
// )

// // SecService ...
// type SecService struct {
// 	pool *pgxpool.Pool
// }

// // SecondService ...
// func SecondService(pool *pgxpool.Pool) *SecService {
// 	return &SecService{pool: pool}
// }

// type clientsData struct {
// 	cLogin	string
// 	cPass	string
// }

// // Auth ...
// func (s *SecService) Auth(cxt context.Context, login, password string) (ok bool) {
// 	cc := &clientsData{}
// 	err := s.pool.QueryRow(cxt, `
// 		SELECT login, password FROM managers WHERE login = $1
// 	`, login).Scan(&cc.cLogin, &cc.cPass)

// 	if err != nil {
// 		log.Print(err)
// 		return false
// 	}

// 	if cc.cPass != password {
// 		log.Println("password is wrong")
// 		return false
// 	}

// 	return true
// }
