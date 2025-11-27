package generator

import (
	"fmt"
)

func (g *Generator) generateDatabasePackages() error {
	if g.config.HasDatabase("postgres") {
		if err := g.generatePostgresDB(); err != nil {
			return err
		}
	}

	if g.config.HasDatabase("mysql") {
		if err := g.generateMySQLDB(); err != nil {
			return err
		}
	}

	if g.config.HasDatabase("mongodb") {
		if err := g.generateMongoDB(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generatePostgresDB() error {
	urlRef := g.getConfigFieldReference("PostgresURL")
	content := fmt.Sprintf(`package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"%s/internal/config"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, cfg *config.Config) (*PostgresDB, error) {
	pool, err := pgxpool.New(ctx, %s)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %%w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %%w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}
`, g.config.ModulePath, urlRef)

	return g.writeFile("internal/database/postgres.go", content)
}

func (g *Generator) generateMySQLDB() error {
	urlRef := g.getConfigFieldReference("MySQLURL")
	content := fmt.Sprintf(`package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"%s/internal/config"
)

type MySQLDB struct {
	db *sql.DB
}

func NewMySQLDB(ctx context.Context, cfg *config.Config) (*MySQLDB, error) {
	db, err := sql.Open("mysql", %s)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %%w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %%w", err)
	}

	return &MySQLDB{db: db}, nil
}

func (db *MySQLDB) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *MySQLDB) DB() *sql.DB {
	return db.db
}
`, g.config.ModulePath, urlRef)

	return g.writeFile("internal/database/mysql.go", content)
}

func (g *Generator) generateMongoDB() error {
	urlRef := g.getConfigFieldReference("MongoURL")
	content := fmt.Sprintf(`package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"%s/internal/config"
)

type MongoDB struct {
	client *mongo.Client
}

func NewMongoDB(ctx context.Context, cfg *config.Config) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(%s)
	
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %%w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %%w", err)
	}

	return &MongoDB{client: client}, nil
}

func (db *MongoDB) Close(ctx context.Context) error {
	if db.client != nil {
		return db.client.Disconnect(ctx)
	}
	return nil
}

func (db *MongoDB) Client() *mongo.Client {
	return db.client
}

func (db *MongoDB) Database(name string) *mongo.Database {
	return db.client.Database(name)
}
`, g.config.ModulePath, urlRef)

	return g.writeFile("internal/database/mongodb.go", content)
}

func (g *Generator) generateCachePackage() error {
	urlRef := g.getConfigFieldReference("RedisURL")
	content := fmt.Sprintf(`package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"%s/internal/config"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(ctx context.Context, cfg *config.Config) (*RedisCache, error) {
	opts, err := redis.ParseURL(%s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %%w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %%w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *RedisCache) Client() *redis.Client {
	return c.client
}
`, g.config.ModulePath, urlRef)

	return g.writeFile("internal/cache/redis.go", content)
}
