package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host           string
	Port           int
	ManagementPort int
	LoggingJSON    bool
{% if has_persistence %}	DatabaseURL    string
{% endif %}{% if has_cache %}	RedisURL       string
{% endif %}{% if messaging == "Kafka" %}	KafkaBrokers      string
	KafkaTopic        string
	KafkaUsername     string
	KafkaPassword     string
	KafkaSASLMechanism string
{% elseif messaging == "Pulsar" %}	PulsarBrokerURL      string
	PulsarTopic          string
	PulsarJWTToken       string
	PulsarSubscriptionName string
{% endif %}{% if has_s3 %}	S3    S3Config
{% endif %}{% if has_azure_blob %}	Azure AzureBlobConfig
{% endif %}}

func Load() (*Config, error) {
	// SERVER_PORT is what the platform manifests inject for HTTP transports.
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "{{ service_port }}"))
	if err != nil {
		return nil, fmt.Errorf("config: SERVER_PORT: %w", err)
	}
	mgmtPort, err := strconv.Atoi(getEnv("MANAGEMENT_PORT", "{{ management_port }}"))
	if err != nil {
		return nil, fmt.Errorf("config: MANAGEMENT_PORT: %w", err)
	}
	cfg := &Config{
		Host:           getEnv("HOST", "0.0.0.0"),
		Port:           port,
		ManagementPort: mgmtPort,
		LoggingJSON:    getEnv("LOGGING_STRUCTURED", "false") == "true",
{% if has_persistence %}		DatabaseURL:    dbURL(),
{% endif %}{% if has_cache %}		RedisURL:       redisURL(),
{% endif %}{% if messaging == "Kafka" %}		KafkaBrokers:       getEnv("MESSAGING_BROKERS", "localhost:9092"),
		KafkaTopic:         getEnv("MESSAGING_TOPIC", "{{ prefix_name }}-{{ suffix_name }}"),
		KafkaUsername:      getEnv("MESSAGING_USERNAME", ""),
		KafkaPassword:      getEnv("MESSAGING_PASSWORD", ""),
		KafkaSASLMechanism: getEnv("MESSAGING_SASL_MECHANISM", ""),
{% elseif messaging == "Pulsar" %}		PulsarBrokerURL:       getEnv("MESSAGING_BROKER_URL", "pulsar://localhost:6650"),
		PulsarTopic:           getEnv("MESSAGING_TOPIC", "{{ prefix_name }}-{{ suffix_name }}"),
		PulsarJWTToken:        getEnv("MESSAGING_JWT_TOKEN", ""),
		PulsarSubscriptionName: getEnv("MESSAGING_SUBSCRIPTION_NAME", "{{ prefix_name }}-{{ suffix_name }}-sub"),
{% endif %}{% if has_s3 %}		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT",   "http://localhost:9000"),
			Bucket:    getEnv("S3_BUCKET",     "{{ project-name }}"),
			Prefix:    getEnv("S3_PREFIX",     ""),
			AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		},
{% endif %}{% if has_azure_blob %}		Azure: AzureBlobConfig{
			Endpoint:    getEnv("AZURE_ENDPOINT",      "http://localhost:10000/devstoreaccount1"),
			Container:   getEnv("AZURE_CONTAINER",    "{{ project-name }}"),
			AccountName: getEnv("AZURE_ACCOUNT_NAME", "devstoreaccount1"),
			AccountKey:  getEnv("AZURE_ACCOUNT_KEY",  "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KkZB2M0XK3Xg=="),
		},
{% endif %}	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

{% if has_persistence %}// dbURL returns DATABASE_URL if set, otherwise assembles it from the individual
// DB_HOST/PORT/USERNAME/PASSWORD/DBNAME env vars that PAO injects from the connection secret.
func dbURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_DBNAME")
	if host != "" && port != "" && user != "" && name != "" {
{% if persistence == "PostgreSQL" %}		return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, pass, host, port, name)
{% elseif persistence == "MySQL" %}		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, name)
{% endif %}	}
{% if persistence == "PostgreSQL" %}	return "postgresql://postgres:postgres@localhost:5432/{{ project-name }}"
{% elseif persistence == "MySQL" %}	return "root:root@tcp(localhost:3306)/{{ project-name }}"
{% endif %}}
{% endif %}{% if has_cache %}// redisURL returns REDIS_URL if set, otherwise assembles it from the individual
// CACHE_HOST/PORT/PASSWORD env vars that PAO injects from the connection secret.
func redisURL() string {
	if v := os.Getenv("REDIS_URL"); v != "" {
		return v
	}
	host := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	pass := os.Getenv("CACHE_PASSWORD")
	if host != "" && port != "" {
		if pass != "" {
			return fmt.Sprintf("redis://:%s@%s:%s", pass, host, port)
		}
		return fmt.Sprintf("redis://%s:%s", host, port)
	}
	return "redis://localhost:6379"
}
{% endif %}
