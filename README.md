# Real-Time Chat Application

## Overview

This project is a real-time chat application built with Go (backend) and Next.js (frontend). It leverages WebSocket connections to provide instant messaging capabilities. The application is designed to run on a single server, demonstrating efficient management of concurrent connections and real-time state updates.

## Key Features

- Real-time messaging using WebSockets
- Go backend for efficient concurrent connection handling
- Next.js frontend for a responsive user interface
- PostgreSQL database for data persistence
- Docker support for easy deployment and development

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.16 or later)
- [Node.js](https://nodejs.org/en/download/) (version 20 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

## Installation

1. Clone the repository:

   ```
   git clone https://github.com/muhreeowki/chat-application.git
   cd chat-application
   ```

2. Set up the environment variables:
   Create a `.env` file in the root directory with the following content:

   ```
   JWT_SECRET="yoursecretkey"
   DB_CONN_STR="user=postgres dbname=postgres password=chat sslmode=disable"
   ```

   Replace `user`, `dbname`, and `password` with your PostgreSQL credentials.

3. If you don't have a PostgreSQL instance running, you can start one using Docker:
   ```
   docker run --name chat-postgres -e POSTGRES_PASSWORD=yourpassword -p 5432:5432 -d postgres
   ```
   Make sure to update the `.env` file with the correct credentials.

## Running the Application

The project uses a Makefile to simplify running both the backend and frontend.

1. Start the backend:

   ```
   make backend
   ```

2. In a new terminal, start the frontend:

   ```
   make client
   ```

3. Access the application by opening a web browser and navigating to `http://localhost:3000`

## Docker Deployment

To deploy the entire application using Docker:

1. Build the Docker images:

   ```
   make docker-build
   ```

2. Start the containers:

   ```
   make docker-up
   ```

3. To stop and remove the containers:
   ```
   make docker-down
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
