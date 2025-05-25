# 💳 Payment Microservice

Um microserviço completo de processamento de pagamentos desenvolvido em Go, simulando os desafios e tecnologias utilizadas.

## 🚀 Funcionalidades

- **gRPC e HTTP APIs** para iniciar pagamentos
- **Validação completa** de dados de cartão e saldo
- **Persistência em PostgreSQL** com status de transação
- **Fila de mensagens Kafka** para processamento assíncrono
- **Logs estruturados** com Logrus
- **Métricas** com Prometheus e OpenTelemetry
- **Workers com goroutines** para processamento paralelo
- **Testes automatizados** com cobertura completa
- **Containerização** com Docker e Docker Compose

## 🏗️ Arquitetura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   gRPC Client   │    │   Prometheus    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Payment Microservice                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Handler   │  │   Service   │  │        Metrics          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Repository  │  │    Queue    │  │        Config           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │      Kafka      │    │      Redis      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 Tecnologias Utilizadas

- **Go 1.21** - Linguagem principal
- **PostgreSQL** - Banco de dados principal
- **Redis** - Cache (opcional)
- **Apache Kafka** - Fila de mensagens
- **Prometheus** - Métricas e monitoramento
- **Grafana** - Visualização de métricas
- **Docker & Docker Compose** - Containerização
- **Gin** - Framework HTTP
- **gRPC** - Comunicação de alta performance
- **Logrus** - Logs estruturados
- **Testify** - Framework de testes

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.21+ (para desenvolvimento local)

### 1. Clone o repositório

```bash
git clone <repository-url>
cd golang-payment-microservice
```

### 2. Execute com Docker Compose

```bash
# Subir todos os serviços
docker-compose up -d

# Verificar logs
docker-compose logs -f payment-service
```

### 3. Verificar se está funcionando

```bash
# Health check
curl http://localhost:8080/health

# Métricas
curl http://localhost:2112/metrics
```

## 📊 Monitoramento

### Prometheus

- **URL**: http://localhost:9090
- **Métricas**: http://localhost:2112/metrics

### Grafana

- **URL**: http://localhost:3000
- **Usuário**: admin
- **Senha**: admin

## 🔧 APIs Disponíveis

### HTTP REST API

#### Criar Pagamento

```bash
POST /api/v1/payments
Content-Type: application/json

{
  "card_number": "1234567890123456",
  "card_holder": "João Silva",
  "expiry_month": 12,
  "expiry_year": 2025,
  "cvv": "123",
  "amount": 100.50,
  "currency": "BRL",
  "merchant_id": "merchant123"
}
```

#### Buscar Pagamento

```bash
GET /api/v1/payments/{payment_id}
```

#### Listar Pagamentos por Merchant

```bash
GET /api/v1/merchants/{merchant_id}/payments?limit=10&offset=0
```

### Exemplos de Uso

```bash
# Criar um pagamento
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "card_number": "1234567890123456",
    "card_holder": "João Silva",
    "expiry_month": 12,
    "expiry_year": 2025,
    "cvv": "123",
    "amount": 100.50,
    "currency": "BRL",
    "merchant_id": "merchant123"
  }'

# Buscar um pagamento
curl http://localhost:8080/api/v1/payments/{payment_id}

# Listar pagamentos de um merchant
curl "http://localhost:8080/api/v1/merchants/merchant123/payments?limit=5&offset=0"
```

## 🗄️ Banco de Dados

### Contas de Teste Disponíveis

| Número do Cartão | Saldo       | Status  |
| ---------------- | ----------- | ------- |
| 1234567890123456 | R$ 1.000,00 | Ativo   |
| 2345678901234567 | R$ 500,00   | Ativo   |
| 3456789012345678 | R$ 2.000,00 | Ativo   |
| 4567890123456789 | R$ 100,00   | Ativo   |
| 5678901234567890 | R$ 0,00     | Ativo   |
| 6789012345678901 | R$ 5.000,00 | Inativo |

### Schema do Banco

```sql
-- Tabela de pagamentos
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    card_number VARCHAR(16) NOT NULL,
    card_holder VARCHAR(100) NOT NULL,
    expiry_month INTEGER NOT NULL,
    expiry_year INTEGER NOT NULL,
    cvv VARCHAR(3) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    merchant_id VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    error_msg TEXT
);

-- Tabela de contas
CREATE TABLE accounts (
    card_number VARCHAR(16) PRIMARY KEY,
    balance DECIMAL(10,2) NOT NULL,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## 📈 Métricas Disponíveis

- `payments_created_total` - Total de pagamentos criados
- `payments_processed_total` - Total de pagamentos processados por status
- `payment_processing_duration_seconds` - Tempo de processamento
- `payment_amount_total` - Valor total de pagamentos
- `http_requests_total` - Total de requisições HTTP
- `http_request_duration_seconds` - Duração das requisições
- `database_connections_active` - Conexões ativas do banco
- `kafka_messages_total` - Total de mensagens Kafka

## 🧪 Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Executar testes específicos
go test ./test -v
```

## 🔧 Desenvolvimento Local

### 1. Instalar dependências

```bash
go mod download
```

### 2. Configurar variáveis de ambiente

```bash
cp .env.example .env
# Editar .env com suas configurações
```

### 3. Executar serviços de dependência

```bash
# Apenas PostgreSQL, Redis e Kafka
docker-compose up -d postgres redis kafka zookeeper
```

### 4. Executar o microserviço

```bash
go run cmd/main.go
```

## 🐳 Docker

### Build da imagem

```bash
docker build -t payment-microservice .
```

### Executar container

```bash
docker run -p 8080:8080 -p 2112:2112 payment-microservice
```

## 📝 Estrutura do Projeto

```
.
├── cmd/
│   └── main.go                 # Ponto de entrada da aplicação
├── internal/
│   ├── handler/               # APIs HTTP e gRPC
│   │   └── http_handler.go
│   ├── service/               # Lógica de negócio
│   │   └── payment_service.go
│   ├── repository/            # Acesso ao banco de dados
│   │   └── payment_repository.go
│   ├── model/                 # Structs e tipos
│   │   └── payment.go
│   ├── queue/                 # Kafka e filas
│   │   ├── kafka_producer.go
│   │   └── kafka_consumer.go
│   └── metrics/               # Métricas Prometheus
│       └── metrics.go
├── config/
│   └── config.go              # Configurações
├── test/
│   └── payment_service_test.go # Testes unitários
├── migrations/
│   └── 001_create_tables.sql  # Migrações do banco
├── docker-compose.yml         # Orquestração de containers
├── Dockerfile                 # Imagem Docker
├── prometheus.yml             # Configuração Prometheus
├── go.mod                     # Dependências Go
└── README.md                  # Documentação
```

## 🔒 Validações Implementadas

### Validação de Cartão

- Número do cartão (16 dígitos)
- Data de expiração (não pode estar vencido)
- CVV (3 dígitos)
- Nome do portador (obrigatório)

### Validação de Pagamento

- Valor maior que zero
- Moeda válida (3 caracteres)
- Merchant ID obrigatório
- Saldo suficiente na conta
- Conta ativa

### Validação de Conta

- Saldo não negativo
- Status ativo/inativo
- Número do cartão único

## 🚦 Status de Pagamento

- `pending` - Pagamento criado, aguardando processamento
- `processing` - Pagamento sendo processado
- `completed` - Pagamento processado com sucesso
- `failed` - Pagamento falhou
- `cancelled` - Pagamento cancelado

## 🔄 Fluxo de Processamento

1. **Recebimento**: API recebe solicitação de pagamento
2. **Validação**: Valida dados do cartão e saldo
3. **Persistência**: Salva pagamento no banco com status `pending`
4. **Enfileiramento**: Envia mensagem para Kafka
5. **Processamento Assíncrono**: Worker processa pagamento
6. **Atualização**: Atualiza status e debita saldo
7. **Métricas**: Registra métricas de sucesso/falha

## 🛠️ Configuração

### Variáveis de Ambiente

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_db
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=payment-processing

# Server
HTTP_PORT=8080
GRPC_PORT=9090
HOST=0.0.0.0

# Metrics
METRICS_PORT=2112
METRICS_PATH=/metrics
```

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 🎯 Próximos Passos

- [ ] Implementar gRPC server
- [ ] Adicionar autenticação JWT
- [ ] Implementar rate limiting
- [ ] Adicionar circuit breaker
- [ ] Implementar retry policies
- [ ] Adicionar mais validações de cartão (Luhn algorithm)
- [ ] Implementar webhooks para notificações
- [ ] Adicionar suporte a múltiplas moedas
- [ ] Implementar reconciliação de pagamentos
- [ ] Adicionar dashboard de monitoramento

## 📞 Suporte

Para dúvidas ou suporte, abra uma issue no repositório ou entre em contato através do email.

---

**Desenvolvido com ❤️ para simular os desafios da CloudWalk**
