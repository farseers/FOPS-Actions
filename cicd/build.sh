export https_proxy=http://192.168.1.88:7890 http_proxy=http://192.168.1.88:7890 all_proxy=socks5://192.168.1.88:7890
docker build -t steden88/cicd:3.0 --network=host --build-arg HTTP_PROXY=http://192.168.1.88:7890 --build-arg HTTPS_PROXY=http://192.168.1.88:7890 -f ./Dockerfile .
docker push steden88/cicd:3.0