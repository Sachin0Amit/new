# Sovereign Intelligence Core - Frontend

Modern Vite + TypeScript frontend for the Sovereign Intelligence Core.

## Architecture

- **Build Tool**: [Vite](https://vitejs.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/) (Strict Mode)
- **State Management**: Class-based modular services
- **Real-time**: Typed WebSockets with exponential backoff
- **Visualization**: Lightweight Charts for financial data

## Development

1.  **Install Dependencies**:
    ```bash
    npm install
    ```

2.  **Start Dev Server**:
    ```bash
    npm run dev
    ```
    This will proxy `/api` and `/ws` to the Go backend at `http://localhost:8081`.

3.  **Build for Production**:
    ```bash
    npm run build
    ```
    Output will be in `web/dist/` with gzip compression enabled.

4.  **Linting**:
    ```bash
    npm run lint
    ```

## Key Modules

- `chat.ts`: WebSocket client with strong message typing.
- `finance.ts`: Candlestick charts and polling fallbacks.
- `settings.ts`: Authenticated preference management with debounced persistence.
- `login.ts`: Secure JWT handling and silent refresh.
