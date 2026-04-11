package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	auth "github.com/f4ke-n0name/autoparts-hub/gen/auth"
	catalog "github.com/f4ke-n0name/autoparts-hub/gen/catalog"
	inventory "github.com/f4ke-n0name/autoparts-hub/gen/inventory"
	order "github.com/f4ke-n0name/autoparts-hub/gen/order"
)

type Clients struct {
	Auth      auth.AuthClient
	Catalog   catalog.CatalogServiceClient
	Order     order.OrderServiceClient
	Inventory inventory.InventoryServiceClient

	conns []*grpc.ClientConn
}

func New(authAddr, catalogAddr, orderAddr, inventoryAddr string) (*Clients, error) {
	authConn, err := dial(authAddr)
	if err != nil {
		return nil, fmt.Errorf("connect to auth: %w", err)
	}

	catalogConn, err := dial(catalogAddr)
	if err != nil {
		return nil, fmt.Errorf("connect to catalog: %w", err)
	}

	orderConn, err := dial(orderAddr)
	if err != nil {
		return nil, fmt.Errorf("connect to order: %w", err)
	}

	inventoryConn, err := dial(inventoryAddr)
	if err != nil {
		return nil, fmt.Errorf("connect to inventory: %w", err)
	}

	return &Clients{
		Auth:      auth.NewAuthClient(authConn),
		Catalog:   catalog.NewCatalogServiceClient(catalogConn),
		Order:     order.NewOrderServiceClient(orderConn),
		Inventory: inventory.NewInventoryServiceClient(inventoryConn),
		conns:     []*grpc.ClientConn{authConn, catalogConn, orderConn, inventoryConn},
	}, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		conn.Close()
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
