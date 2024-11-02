# Image Fetch Service

This service allows you to proxy images by passing on base64-encoded url as query. It fetches the image from the URL and returns it as a response. If the image is not found, it returns a fallback image.

## Prerequisites

Make sure you have Docker and Docker Compose installed on your machine. You can download them from the official Docker website.

## Getting Started with Docker

1. **Pull the Image**  
   To start using the image-fetch service, pull the latest image from the Docker registry:
   ```bash
   docker pull ghcr.io/digital39999/img-fetch:latest
   ```

2. **Run the Container**  
   You can run the container using the following command:
   ```bash
   docker run -e FALLBACK_IMAGE_URL="http://fallback.image.url/image.jpg" -e PORT=8080 -p 8080:8080 ghcr.io/digital39999/img-fetch
   ```

3. **Access the Service**  
   Once the container is running, you can access the service at `http://localhost:8080/image`. Use a query parameter `hash` to pass in the base64-encoded URL.

## Running with Docker Compose

If you prefer to use Docker Compose, follow these steps:

1. **Create a `docker-compose.yml` File**  
   Here’s an example of a `docker-compose.yml` that includes both the image-fetch service and a Redis Scheduler:

   ```yaml
   version: '3.8'

   services:
     image-fetch:
       image: ghcr.io/digital39999/img-fetch:latest
       environment:
         FALLBACK_IMAGE_URL: "http://fallback.image.url/image.jpg"
         PORT: 8080
       ports:
         - "8080:8080"
   ```

2. **Run the Services**  
   Navigate to the directory containing your `docker-compose.yml` file and run:
   ```bash
   docker-compose up -d
   ```

3. **Access the Service**  
   Once the container is running, you can access the service at `http://localhost:8080/image`. Use a query parameter `hash` to pass in the base64-encoded URL.

## Conclusion

You can now use the image-fetch service with Docker and Docker Compose to shorten and fetch images easily. For further details or custom configurations, feel free to modify the Docker Compose settings or environment variables as needed.
