# 🏠 Cupboard Inventory

A lightweight web application for tracking pantry items and freezer meals — know what you have and reduce waste.

## Features

- **Pantry tracker** — add, edit and delete pantry items with name, quantity, category, expiry date and notes; items are sorted by nearest expiry first
- **Freezer meals tracker** — log leftover meals with portions and freeze date; oldest meals are surfaced first so nothing gets forgotten
- Expiry warnings (expired / expiring within 7 days) highlighted on pantry cards
- All data persisted locally in a `data.json` file — no database required

## Prerequisites

- [Go](https://go.dev/dl/) 1.21 or newer

## Running Locally

```bash
# Clone the repository
git clone https://github.com/JockByTheSea/cupboard-inventory.git
cd cupboard-inventory

# Start the server (default port 8080)
go run .
```

Then open <http://localhost:8080> in your browser.

To use a different port, set the `PORT` environment variable before starting:

```bash
PORT=9090 go run .
```

### Building a Binary

```bash
go build -o cupboard-inventory .
./cupboard-inventory
```

## Project Structure

```
.
├── main.go          # Route registration and server startup
├── models.go        # Data types: PantryItem, FreezerMeal, Store
├── store.go         # JSON persistence: loadStore / saveStore
├── templates.go     # Template helpers (funcMap) and initialisation
├── handlers.go      # HTTP handlers for all routes
├── static/
│   └── style.css    # Application stylesheet
└── templates/
    └── index.html   # Main HTML template
```

Data is stored at runtime in `data.json` in the working directory (excluded from version control).

## Contributing

1. Fork the repository and create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes, following the existing code style.
3. Verify the build is clean before opening a pull request:
   ```bash
   go build ./...
   go vet ./...
   ```
4. Open a pull request against `main` with a clear description of what was changed and why.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
