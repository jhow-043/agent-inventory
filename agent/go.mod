module inventario/agent

go 1.24.0

replace inventario/shared => ../shared

require (
	github.com/yusufpapurcu/wmi v1.2.4
	golang.org/x/sys v0.41.0
	inventario/shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/uuid v1.6.0 // indirect
)
