module {{ module_path }}

go 1.23

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/prometheus/client_golang v1.20.5
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0
	go.opentelemetry.io/otel/sdk v1.33.0
{% if persistence == "PostgreSQL" %}	github.com/jackc/pgx/v5 v5.7.2
{% elseif persistence == "MySQL" %}	github.com/go-sql-driver/mysql v1.8.1
{% endif %}{% if has_cache %}	github.com/redis/go-redis/v9 v9.7.3
{% endif %}{% if messaging == "Kafka" %}	github.com/IBM/sarama v1.44.0
	github.com/xdg-go/scram v1.1.2
{% elseif messaging == "Pulsar" %}	github.com/apache/pulsar-client-go v0.14.0
{% endif %}{% if has_s3 %}	github.com/aws/aws-sdk-go-v2 v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.26.0
	github.com/aws/aws-sdk-go-v2/credentials v1.16.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.47.0
{% endif %}{% if has_azure_blob %}	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.0
{% endif %})
