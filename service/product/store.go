package product

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/thomzes/ecom-go/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetProducts() ([]*types.Product, error) {
	rows, err := s.db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}

	products := make([]*types.Product, 0)
	for rows.Next() {
		p, err := scanRowIntoProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}

func scanRowIntoProduct(rows *sql.Rows) (*types.Product, error) {
	product := new(types.Product)
	err := rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Image,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *Store) GetProductByID(id int) (*types.Product, error) {
	rows, err := s.db.Query("SELECT * FROM products WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	p := new(types.Product)
	for rows.Next() {
		p, err = scanRowIntoProduct(rows)
		if err != nil {
			return nil, err
		}
	}

	if p.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}

	return p, nil
}

func (s *Store) GetProductsByID(productsIDs []int) ([]types.Product, error) {
	if len(productsIDs) == 0 {
		return []types.Product{}, nil
	}

	inPlaceholders := "?"
	if len(productsIDs) > 1 {
		inPlaceholders += strings.Repeat(",?", len(productsIDs)-1)
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, image, price, quantity, createdAt
		FROM products
		WHERE id IN (%s)
	`, inPlaceholders)

	// Args for IN clause + Args for FIELD() to preserve input order
	args := make([]interface{}, len(productsIDs))
	for i, id := range productsIDs {
		args[i] = id
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]types.Product, 0, len(productsIDs))
	for rows.Next() {
		p, err := scanRowIntoProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, *p)
	}

	return products, nil

}

func (s *Store) CreateProduct(product types.CreateProductPayload) error {
	_, err := s.db.Exec("INSERT INTO products (name, description, image, price, quantity) VALUES (?,?,?,?,?)", product.Name, product.Description, product.Image, product.Price, product.Quantity)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateProduct(product types.Product) error {
	_, err := s.GetProductByID(product.ID)
	if err != nil {
		return fmt.Errorf("product with id %d not found", product.ID)
	}

	result, err := s.db.Exec("UPDATE products SET name = ?, description = ?, image = ?, price = ?, quantity = ? WHERE id = ?", product.Name, product.Description, product.Image, product.Price, product.Quantity, product.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no product updated")
	}

	return nil
}

func (s *Store) DeleteProduct(id int) error {
	_, err := s.GetProductByID(id)
	if err != nil {
		return fmt.Errorf("product with id %d, not found", id)
	}

	result, err := s.db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no product deleted")
	}

	return nil
}
