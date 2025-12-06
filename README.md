# Loam 🌱

> A Transactional Storage Engine for Markdown + Frontmatter backed by Git.

**Loam** trata seu diretório de notas Markdown como um banco de dados NoSQL.
Ele oferece operações de CRUD atômicas e seguras, garantindo que suas automações não corrompam seu cofre pessoal.

## 🚀 Instalação

```bash
go install github.com/aretw0/loam/cmd/loam@latest
```

## 🛠️ Uso

### Inicializar um Cofre

```bash
mkdir notas
cd notas
loam init
```

### Criar/Editar Nota

```bash
loam write -id minha-nota -content "Texto da nota"
```

### Salvar (Commit)

```bash
loam commit -m "Minha primeira nota"
```

### Ler Nota (Raw)

```bash
loam read -id minha-nota
```

## 📚 Documentação

- [Visão do Produto](docs/PRODUCT.md)
- [Arquitetura Técnica](docs/TECHNICAL.md)
- [Roadmap](docs/PLANNING.md)

## Status

🚧 **Alpha**. O kernel e a CLI básica estão funcionais, mas a API pode mudar.
Use por sua conta e risco (mas hey, é Git, você pode sempre dar revert!).
