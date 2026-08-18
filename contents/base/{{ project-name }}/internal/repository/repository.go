// Package repository persists the service's entity, {{ EntityName }} — the same five CRUD
// operations every p6m service transport serves over `{ id, display_name }`.
package repository

{% if persistence == "PostgreSQL" %}
import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"{{ module_path }}/internal/persistence"
)
{% elseif persistence == "MySQL" %}
import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"{{ module_path }}/internal/persistence"
)
{% else %}
import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
)
{% endif %}

// {{ EntityName }} is the entity this service serves.
type {{ EntityName }} struct {
	ID          string
	DisplayName string
}

// ErrNotFound reports an id no stored {{ EntityName }} backs.
var ErrNotFound = errors.New("{{ entity-name }} not found")

// newID mints a random UUIDv4 without an external dependency.
func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // the platform CSPRNG failing is unrecoverable
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

{% if persistence == "PostgreSQL" %}
// Store persists {{ EntityName }}s in PostgreSQL through the shared connection pool.
type Store struct{}

// New ensures the backing table exists and returns the store.
func New(ctx context.Context) (*Store, error) {
	_, err := persistence.DB().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS {{ entity_name }}s (
			id           TEXT PRIMARY KEY,
			display_name TEXT NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("repository: ensure schema: %w", err)
	}
	return &Store{}, nil
}

func (s *Store) Create(ctx context.Context, displayName string) ({{ EntityName }}, error) {
	e := {{ EntityName }}{ID: newID(), DisplayName: displayName}
	_, err := persistence.DB().Exec(ctx,
		`INSERT INTO {{ entity_name }}s (id, display_name) VALUES ($1, $2)`, e.ID, e.DisplayName)
	return e, err
}

func (s *Store) Get(ctx context.Context, id string) ({{ EntityName }}, error) {
	var e {{ EntityName }}
	err := persistence.DB().QueryRow(ctx,
		`SELECT id, display_name FROM {{ entity_name }}s WHERE id = $1`, id).Scan(&e.ID, &e.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (s *Store) List(ctx context.Context) ([]{{ EntityName }}, error) {
	rows, err := persistence.DB().Query(ctx,
		`SELECT id, display_name FROM {{ entity_name }}s ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []{{ EntityName }}{}
	for rows.Next() {
		var e {{ EntityName }}
		if err := rows.Scan(&e.ID, &e.DisplayName); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) Update(ctx context.Context, id, displayName string) ({{ EntityName }}, error) {
	tag, err := persistence.DB().Exec(ctx,
		`UPDATE {{ entity_name }}s SET display_name = $2 WHERE id = $1`, id, displayName)
	if err != nil {
		return {{ EntityName }}{}, err
	}
	if tag.RowsAffected() == 0 {
		return {{ EntityName }}{}, ErrNotFound
	}
	return {{ EntityName }}{ID: id, DisplayName: displayName}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	tag, err := persistence.DB().Exec(ctx,
		`DELETE FROM {{ entity_name }}s WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
{% elseif persistence == "MySQL" %}
// Store persists {{ EntityName }}s in MySQL through the shared connection.
type Store struct{}

// New ensures the backing table exists and returns the store.
func New(ctx context.Context) (*Store, error) {
	_, err := persistence.DB().ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS {{ entity_name }}s (
			id           VARCHAR(36) PRIMARY KEY,
			display_name VARCHAR(255) NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("repository: ensure schema: %w", err)
	}
	return &Store{}, nil
}

func (s *Store) Create(ctx context.Context, displayName string) ({{ EntityName }}, error) {
	e := {{ EntityName }}{ID: newID(), DisplayName: displayName}
	_, err := persistence.DB().ExecContext(ctx,
		`INSERT INTO {{ entity_name }}s (id, display_name) VALUES (?, ?)`, e.ID, e.DisplayName)
	return e, err
}

func (s *Store) Get(ctx context.Context, id string) ({{ EntityName }}, error) {
	var e {{ EntityName }}
	err := persistence.DB().QueryRowContext(ctx,
		`SELECT id, display_name FROM {{ entity_name }}s WHERE id = ?`, id).Scan(&e.ID, &e.DisplayName)
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (s *Store) List(ctx context.Context) ([]{{ EntityName }}, error) {
	rows, err := persistence.DB().QueryContext(ctx,
		`SELECT id, display_name FROM {{ entity_name }}s ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []{{ EntityName }}{}
	for rows.Next() {
		var e {{ EntityName }}
		if err := rows.Scan(&e.ID, &e.DisplayName); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) Update(ctx context.Context, id, displayName string) ({{ EntityName }}, error) {
	res, err := persistence.DB().ExecContext(ctx,
		`UPDATE {{ entity_name }}s SET display_name = ? WHERE id = ?`, displayName, id)
	if err != nil {
		return {{ EntityName }}{}, err
	}
	if n, err := res.RowsAffected(); err != nil {
		return {{ EntityName }}{}, err
	} else if n == 0 {
		return {{ EntityName }}{}, ErrNotFound
	}
	return {{ EntityName }}{ID: id, DisplayName: displayName}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := persistence.DB().ExecContext(ctx,
		`DELETE FROM {{ entity_name }}s WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}
{% else %}
// Store keeps {{ EntityName }}s in memory — swap in a persistence selection to back it with a
// real database.
type Store struct {
	mu    sync.Mutex
	items map[string]{{ EntityName }}
	order []string
}

// New returns an empty in-memory store.
func New(ctx context.Context) (*Store, error) {
	return &Store{items: map[string]{{ EntityName }}{}}, nil
}

func (s *Store) Create(ctx context.Context, displayName string) ({{ EntityName }}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := {{ EntityName }}{ID: newID(), DisplayName: displayName}
	s.items[e.ID] = e
	s.order = append(s.order, e.ID)
	return e, nil
}

func (s *Store) Get(ctx context.Context, id string) ({{ EntityName }}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[id]
	if !ok {
		return {{ EntityName }}{}, ErrNotFound
	}
	return e, nil
}

func (s *Store) List(ctx context.Context) ([]{{ EntityName }}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := []{{ EntityName }}{}
	for _, id := range s.order {
		items = append(items, s.items[id])
	}
	return items, nil
}

func (s *Store) Update(ctx context.Context, id, displayName string) ({{ EntityName }}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[id]
	if !ok {
		return {{ EntityName }}{}, ErrNotFound
	}
	e.DisplayName = displayName
	s.items[id] = e
	return e, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return nil
}
{% endif %}
