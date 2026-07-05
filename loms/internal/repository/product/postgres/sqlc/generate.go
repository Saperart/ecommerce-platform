package sqlc

//go:generate mockgen -source=db.go -destination=mocks/dbtx_mock.go -package=mocks
//go:generate mockgen -destination=mocks/pgx_mock.go -package=mocks github.com/jackc/pgx/v5 Row,Rows
