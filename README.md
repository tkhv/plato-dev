# plato-dev

To run the WhisperLiveServer, cd into WhisperLiveServer and run:

#### GPU
``` 
docker build -f Dockerfile.gpu -t whisperlive-gpu .
docker run -it --gpus all -p 9090:9090 --rm whisperlive-gpu
```
#### Tensorrt
``` 
docker build -f Dockerfile.tensorrt -t whisperlive-tensorrt .
docker run -it --gpus all -p 9090:9090 --rm whisperlive-tensorrt
```
#### CPU
``` 
docker build -f Dockerfile.cpu -t whisperlive-cpu .
docker run -it -p 9090:9090 --rm whisperlive-cpu
```

To run an example app that uses the WhisperLiveServer to transcribe, cd into app and run app.py.