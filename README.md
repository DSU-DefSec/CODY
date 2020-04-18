# maizenet

```
DOCKER_ID=$(sudo docker run -v $(pwd):/opt/maizenet -p 8080:8080 -td ubuntu)
sudo docker exec -it $DOCKER_ID "/bin/bash"
cd /opt/maizenet; ./install.sh
```
