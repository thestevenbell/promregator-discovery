# Promregator-discovery

## Description
This a simple GoLang script to call the Promregator /discovery API, do a basic validation of the return object and then save it to disk.  It is intended to be used with a Dockerized deployment of Promregator and Prometheus where Prometheus is configured to read the targets provided by the Promregator /discovery API and the Prometheus and Promregator images share a persistent Docker volume.  

A Docker image is provided that will both build and allow execution of the binary.  

## Local Development
From the main directory:  
**Build**  
```go build .```  
**Execute**  
```go run main.go -targetUrl=http://localhost:8080/discovery -interval=30 -fileDestination=./promregator-discovery.json```


## Docker
**build the image**  
```docker build -t thestevenbell/promregator-discovery:latest .```  
**remove those pesky unnecessary intermediate images created in the multibuild Docker process**  
```docker image prune --filter label=stage=intermediate```

**run the container for local development on Mac**  

```docker run -it --rm \
--mount type=volume,source=promregator_discovery,target=/promregator_discovery \
--name promregator-discovery thestevenbell/promregator-discovery:latest \
-targetUrl=http://host.docker.internal:8080/discovery \
-interval=10 \
-fileDestination=/promregator_discovery/promregator_discovery.json
```

**run a docker stack with Promregator-Discovery, Promregator and Prometheus.**  
See the [**prometheus.yml**](stack/prometheus.yml) file for example scrape configuration making use of the discovered targets file created by this project.  
```bash
docker stack deploy \
--compose-file stack/docker-compose.yml \
monitoring
```