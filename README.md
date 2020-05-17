# hackernet

features lol


dev environment
```
DOCKER_ID=$(sudo docker run -v $(pwd):/opt/hackernet -p 8080:8080 -td ubuntu)
sudo docker exec -it $DOCKER_ID "/bin/bash"
cd /opt/hackernet && ./install.sh
```

### API

Most of the web interface is read only (with the exception of the ability to join competitions, to deploy labs. This is because I don't want to write forms.

### Events

- `01` competitions
    - `koth`
    - `ctf`
- `10` learn

100 +
- `101`: lesson lol
