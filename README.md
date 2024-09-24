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
   git clone https://github.com/muhreeowki/chat-app.git
   cd chat-app
   ```

## Running the Application on Docker

To start the entire application using Docker:

1. Build and Start the containers:

   ```
   make docker-up
   ```
   This will start the containers and have them running in the background.
   Access the application by opening a web browser and navigating to `http://localhost:3000`


2. To stop and remove the containers:
   ```
   make docker-down
   ```
   
3. To remove the installed images form your system:
   ```
   make docker-rmi
   ```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
