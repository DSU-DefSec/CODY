# Center for Organized DefSec Yuppies

This web application allows students to easily deploy vApps to themselves and to browse DefSec lectures.


```
DOCKER_ID=$(sudo docker run -v $(pwd):/opt/CODY -p 8080:8080 -td ubuntu)
sudo docker exec -it $DOCKER_ID "/bin/bash"
cd /opt/CODY && ./install.sh
```

### Adding Lessons

Most of the web interface is read only (with the exception of the ability to join competitions, to deploy labs. This is because I don't want to write forms.
